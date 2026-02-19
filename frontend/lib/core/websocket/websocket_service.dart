import 'dart:async';
import 'dart:convert';
import 'package:get/get.dart';
import 'package:web_socket_channel/web_socket_channel.dart';

import '../database/database_service.dart';
import '../database/models/message_model.dart';
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
      final event = WSEvent.fromJson(jsonData);

      // Handle different event types
      switch (event.type) {
        case 'new_message':
        case 'message':
          _handleNewMessage(event.data);
          break;
        case 'message_ack':
          _handleMessageAck(event.data);
          break;
        case 'pong':
          // Server acknowledged ping
          break;
        default:
          // Broadcast to other listeners
          _eventController.add(event);
      }
    } catch (e) {
      // Ignore parse errors
    }
  }

  /// Handle new message from server
  void _handleNewMessage(Map<String, dynamic> data) async {
    try {
      final message = MessageModel.fromServerJson(data);
      await DatabaseService.to.saveMessage(message);

      // Update chat session
      await DatabaseService.to.updateChatSessionLastMessage(
        message.chatId,
        message.content ?? '[Media]',
        message.createdAt,
      );
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
        final serverSeqId = data['seq_id'] as int?;
        await DatabaseService.to.updateMessageStatus(
          localId,
          status == 'sent' ? MessageStatus.sent : MessageStatus.failed,
          serverSeqId: serverSeqId,
        );
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
