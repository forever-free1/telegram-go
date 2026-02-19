import 'dart:async';
import 'package:get/get.dart';
import 'package:dio/dio.dart';

import '../database/database_service.dart';
import '../database/models/message_model.dart';
import '../database/models/chat_session_model.dart';
import '../network/api_client.dart';

/// Sync controller - Incremental sync engine
class SyncController extends GetxController {
  final DatabaseService _db = DatabaseService.to;
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
  }

  @override
  void onClose() {
    _periodicSyncTimer?.cancel();
    super.onClose();
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
      // 1. Get local max seqId
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

      // 3. Parse and save messages
      final messagesData = data['messages'] as List<dynamic>? ?? [];
      if (messagesData.isNotEmpty) {
        final messages = messagesData
            .map((json) => MessageModel.fromServerJson(json as Map<String, dynamic>))
            .toList();

        await _db.saveMessages(messages);

        // 4. Update chat sessions
        await _updateChatSessions(messages);
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

  /// Update chat sessions with new messages
  Future<void> _updateChatSessions(List<MessageModel> messages) async {
    // Group messages by chatId
    final Map<int, List<MessageModel>> messagesByChat = {};
    for (final msg in messages) {
      messagesByChat.putIfAbsent(msg.chatId, () => []).add(msg);
    }

    // Update each chat session
    for (final entry in messagesByChat.entries) {
      final chatId = entry.key;
      final chatMessages = entry.value;

      // Get latest message
      final latestMsg = chatMessages.reduce(
        (a, b) => a.createdAt.isAfter(b.createdAt) ? a : b,
      );

      // Get or create chat session
      var session = await _db.getChatSession(chatId);
      session ??= ChatSessionModel(
        chatId: chatId,
        name: 'Chat $chatId',
        unreadCount: 0,
        updatedAt: DateTime.now(),
      );

      // Update last message
      session.updateFromMessage(
        latestMsg.content ?? '[Media]',
        latestMsg.createdAt,
      );

      await _db.saveChatSession(session);
    }
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
