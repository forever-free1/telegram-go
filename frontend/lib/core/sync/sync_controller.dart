import 'dart:async';
import 'package:drift/drift.dart';
import 'package:get/get.dart' hide Value;
import 'package:dio/dio.dart';

import '../database/database.dart';
import '../network/api_client.dart';
import '../websocket/websocket_service.dart';

/// Sync controller - Incremental sync engine using Drift
class SyncController extends GetxController {
  final AppDatabase _db = Get.find<AppDatabase>();
  final ApiClient _api = ApiClient.to;

  final RxBool isSyncing = false.obs;
  final Rx<DateTime?> lastSyncTime = Rx<DateTime?>(null);
  final RxString errorMessage = ''.obs;

  Timer? _periodicSyncTimer;

  @override
  void onInit() {
    super.onInit();
    // Start periodic sync when app is active
    _startPeriodicSync();
    // Connect WebSocket
    _connectWebSocket();
  }

  @override
  void onClose() {
    _periodicSyncTimer?.cancel();
    super.onClose();
  }

  /// Connect to WebSocket
  void _connectWebSocket() {
    final ws = WebSocketService.to;
    if (!ws.isConnected) {
      ws.connect();
    }
  }

  /// Start periodic sync every 30 seconds
  void _startPeriodicSync() {
    _periodicSyncTimer?.cancel();
    _periodicSyncTimer = Timer.periodic(
      const Duration(seconds: 30),
      (_) => syncMessages(),
    );
  }

  /// Manual sync trigger
  Future<void> syncMessages() async {
    if (isSyncing.value) return;

    isSyncing.value = true;
    errorMessage.value = '';

    try {
      // 1. Get local max seqId from Drift
      final maxSeqId = await _db.getMaxSeqId();

      // 2. Call backend sync API
      final response = await _api.get(
        '/sync',
        queryParameters: {'last_seq_id': maxSeqId},
      );

      final responseData = response.data;
      if (responseData is! Map<String, dynamic>) {
        throw Exception('Invalid response format');
      }

      // Extract data from response
      Map<String, dynamic> data;
      if (responseData.containsKey('data')) {
        data = responseData['data'] ?? {};
      } else {
        data = responseData;
      }

      // 3. Parse and save messages using Drift transaction
      final messagesData = data['messages'] as List<dynamic>? ?? [];
      if (messagesData.isNotEmpty) {
        await _saveMessagesAndUpdateChats(messagesData);
      }

      lastSyncTime.value = DateTime.now();
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        // Auth error - will be handled by interceptor
        return;
      }
      errorMessage.value = 'Sync failed: ${e.message}';
    } catch (e) {
      errorMessage.value = 'Sync failed: $e';
    } finally {
      isSyncing.value = false;
    }
  }

  /// Save messages and update chat sessions in a Drift transaction
  Future<void> _saveMessagesAndUpdateChats(List<dynamic> messagesData) async {
    await _db.transaction(() async {
      // Map to track chat sessions that need updating
      final Map<int, _LatestMessageInfo> latestMessagesByChat = {};

      // Insert all messages
      for (final json in messagesData) {
        final msgData = json as Map<String, dynamic>;

        // Handle timestamp - backend sends 'created_at' for sync
        DateTime? createdAt;
        if (msgData['created_at'] != null) {
          createdAt = DateTime.tryParse(msgData['created_at'] as String);
        }
        createdAt ??= DateTime.now();

        final message = MessagesCompanion(
          seqId: Value(msgData['seq_id'] as int? ?? 0),
          chatId: Value(msgData['chat_id'] as int),
          senderId: Value(msgData['sender_id'] as int),
          type: Value(msgData['type'] as int? ?? 1),
          content: Value(msgData['content'] as String?),
          status: const Value('sent'),
          createdAt: Value(createdAt),
        );

        await _db.insertMessage(message);

        // Track latest message per chat
        final chatId = msgData['chat_id'] as int;
        final content = message.content.value ?? '[Media]';

        if (!latestMessagesByChat.containsKey(chatId) ||
            latestMessagesByChat[chatId]!.createdAt.isBefore(createdAt)) {
          latestMessagesByChat[chatId] = _LatestMessageInfo(content, createdAt);
        }
      }

      // Update chat sessions
      for (final entry in latestMessagesByChat.entries) {
        final chatId = entry.key;
        final info = entry.value;

        // Check if chat session exists
        final existingSession = await (_db.select(_db.chatSessions)
              ..where((t) => t.chatId.equals(chatId)))
            .getSingleOrNull();

        if (existingSession != null) {
          // Update existing session
          await (_db.update(_db.chatSessions)
                ..where((t) => t.chatId.equals(chatId)))
              .write(ChatSessionsCompanion(
                lastMessage: Value(info.content),
                updatedAt: Value(info.createdAt),
              ));
        } else {
          // Create new session (use chat ID as name placeholder)
          await _db.insertOrUpdateChat(ChatSessionsCompanion(
            chatId: Value(chatId),
            name: Value('Chat $chatId'),
            lastMessage: Value(info.content),
            unreadCount: const Value(0),
            updatedAt: Value(info.createdAt),
          ));
        }
      }
    });
  }

  /// Initial full sync (when user first logs in)
  Future<void> initialSync() async {
    // Clear local data first
    await _db.clearAll();

    // Do full sync
    await syncMessages();
  }

  /// Stop periodic sync
  void stopPeriodicSync() {
    _periodicSyncTimer?.cancel();
    _periodicSyncTimer = null;
  }

  /// Resume periodic sync
  void resumePeriodicSync() {
    if (_periodicSyncTimer == null) {
      _startPeriodicSync();
    }
  }
}

/// Helper class to track latest message info per chat
class _LatestMessageInfo {
  final String content;
  final DateTime createdAt;

  _LatestMessageInfo(this.content, this.createdAt);
}
