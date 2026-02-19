import 'dart:async';
import 'dart:convert';
import 'package:drift/drift.dart';
import 'package:get/get.dart' hide Value;
import 'package:web_socket_channel/web_socket_channel.dart';

import '../database/database.dart';
import '../network/api_client.dart';

/// WebSocket event types
class WSEvent {
  final String type;
  final Map<String, dynamic> data;

  WSEvent({required this.type, required this.data});

  factory WSEvent.fromJson(Map<String, dynamic> json) {
    return WSEvent(
      type: json['type'] ?? '',
      data: json['data'] ?? {},
    );
  }
}

/// WebSocket service for real-time messaging
class WebSocketService extends GetxService {
  static WebSocketService get to => Get.find<WebSocketService>();

  static const String _wsUrl = 'ws://10.0.2.2:8080/ws';

  WebSocketChannel? _channel;
  Timer? _reconnectTimer;
  Timer? _pingTimer;
  bool _isConnected = false;
  int _reconnectAttempts = 0;
  static const int _maxReconnectAttempts = 5;
  static const int _baseReconnectDelay = 1000; // 1 second

  final _eventController = StreamController<WSEvent>.broadcast();
  Stream<WSEvent> get eventStream => _eventController.stream;

  bool get isConnected => _isConnected;

  Future<WebSocketService> init() async {
    return this;
  }

  /// Connect to WebSocket server
  Future<void> connect() async {
    if (_isConnected) return;

    try {
      final token = await ApiClient.getToken();
      if (token == null || token.isEmpty) {
        return;
      }

      // Append token as query parameter
      final uri = Uri.parse('$_wsUrl?token=$token');
      _channel = WebSocketChannel.connect(uri);

      _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
      );

      _isConnected = true;
      _reconnectAttempts = 0;
      _startPingTimer();
    } catch (e) {
      _isConnected = false;
      _scheduleReconnect();
    }
  }

  /// Disconnect from WebSocket server
  void disconnect() {
    _stopTimers();
    _channel?.sink.close();
    _channel = null;
    _isConnected = false;
  }

  /// Send message through WebSocket
  void send(String eventType, Map<String, dynamic> data) {
    if (!_isConnected || _channel == null) return;

    final message = json.encode({
      'event': eventType,
      'data': data,
    });

    _channel!.sink.add(message);
  }

  /// Handle incoming message
  void _onMessage(dynamic message) {
    try {
      final jsonData = json.decode(message as String) as Map<String, dynamic>;

      // Handle different event types directly (backend sends flat JSON)
      final type = jsonData['type'] as String? ?? '';

      switch (type) {
        case 'new_message':
        case 'message':
          _handleNewMessage(jsonData);
          break;
        case 'WS_MSG_READ':
        case 'message_ack':
          _handleMessageAck(jsonData);
          break;
        case 'pong':
          // Server acknowledged ping
          break;
        default:
          // Broadcast to other listeners
          final event = WSEvent.fromJson(jsonData);
          _eventController.add(event);
      }
    } catch (e) {
      // Ignore parse errors
    }
  }

  /// Handle new message from server
  void _handleNewMessage(Map<String, dynamic> data) async {
    try {
      final db = Get.find<AppDatabase>();

      // Parse timestamp - backend sends 'timestamp', sync might send 'created_at'
      DateTime? createdAt;
      if (data['timestamp'] != null) {
        createdAt = DateTime.tryParse(data['timestamp'] as String);
      } else if (data['created_at'] != null) {
        createdAt = DateTime.tryParse(data['created_at'] as String);
      }
      createdAt ??= DateTime.now();

      // Parse message from server
      final message = MessagesCompanion(
        seqId: Value(data['seq_id'] as int? ?? data['seqId'] as int? ?? 0),
        chatId: Value(data['chat_id'] as int),
        senderId: Value(data['sender_id'] as int),
        type: Value(data['type'] as int? ?? 1),
        content: Value(data['content'] as String?),
        status: const Value('sent'),
        createdAt: Value(createdAt),
      );

      // Save message to Drift
      await db.insertMessage(message);

      // Update chat session
      final sessions = await (db.select(db.chatSessions)
            ..where((t) => t.chatId.equals(data['chat_id'] as int)))
          .get();

      if (sessions.isNotEmpty) {
        await (db.update(db.chatSessions)
              ..where((t) => t.chatId.equals(data['chat_id'] as int)))
            .write(ChatSessionsCompanion(
          lastMessage: Value(data['content'] as String? ?? '[Media]'),
          updatedAt: Value(message.createdAt.value),
        ));
      } else {
        // Create new chat session
        await db.insertOrUpdateChat(ChatSessionsCompanion(
          chatId: Value(data['chat_id'] as int),
          name: Value('Chat ${data['chat_id']}'),
          lastMessage: Value(data['content'] as String? ?? '[Media]'),
          unreadCount: const Value(1),
          updatedAt: Value(DateTime.now()),
        ));
      }
    } catch (e) {
      // Handle error silently
    }
  }

  /// Handle message acknowledgment
  void _handleMessageAck(Map<String, dynamic> data) async {
    try {
      final localId = data['local_id'] as String?;
      final status = data['status'] as String? ?? 'sent';

      if (localId != null) {
        final db = Get.find<AppDatabase>();
        // Find message by content and status (we used content as identifier for local messages)
        final content = data['content'] as String?;
        final chatId = data['chat_id'] as int?;

        if (content != null && chatId != null) {
          final messages = await db.getMessagesByChatId(chatId);
          final targetMsg = messages.where((m) =>
            m.content == content && m.status == 'sending'
          ).toList();

          if (targetMsg.isNotEmpty) {
            final serverSeqId = data['seq_id'] as int?;
            final newStatus = status == 'sent' ? 'sent' : 'failed';

            await db.updateMessageStatus(targetMsg.first.id, newStatus);

            if (serverSeqId != null && serverSeqId > 0) {
              await (db.update(db.messages)..where((t) => t.id.equals(targetMsg.first.id)))
                  .write(MessagesCompanion(seqId: Value(serverSeqId)));
            }
          }
        }
      }
    } catch (e) {
      // Handle error silently
    }
  }

  void _onError(dynamic error) {
    _isConnected = false;
    _scheduleReconnect();
  }

  void _onDone() {
    _isConnected = false;
    _stopTimers();
    _scheduleReconnect();
  }

  /// Schedule reconnect with exponential backoff
  void _scheduleReconnect() {
    if (_reconnectAttempts >= _maxReconnectAttempts) {
      return;
    }

    _reconnectTimer?.cancel();

    // Exponential backoff: 1s, 2s, 4s, 8s, 16s
    final delay = _baseReconnectDelay * (1 << _reconnectAttempts);
    _reconnectAttempts++;

    _reconnectTimer = Timer(Duration(milliseconds: delay), () {
      connect();
    });
  }

  /// Start ping timer to keep connection alive
  void _startPingTimer() {
    _pingTimer?.cancel();
    _pingTimer = Timer.periodic(const Duration(seconds: 30), (_) {
      send('ping', {});
    });
  }

  void _stopTimers() {
    _reconnectTimer?.cancel();
    _pingTimer?.cancel();
  }

  @override
  void onClose() {
    disconnect();
    _eventController.close();
    super.onClose();
  }
}
