import 'dart:async';
import 'package:drift/drift.dart' hide Column;
import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:get/get.dart' hide Value;
import 'package:dio/dio.dart';

import '../../../core/database/database.dart';
import '../../../core/database/models/message_model.dart';
import '../../../core/network/api_client.dart';
import '../../../core/websocket/websocket_service.dart';
import '../../auth/controllers/auth_controller.dart';
import '../widgets/message_bubble.dart';
import '../widgets/chat_input_bar.dart';

/// Chat page screen - real-time messaging with Drift
class ChatPageScreen extends StatefulWidget {
  final int chatId;
  final String chatName;
  final String avatarText;

  const ChatPageScreen({
    super.key,
    required this.chatId,
    required this.chatName,
    required this.avatarText,
  });

  @override
  State<ChatPageScreen> createState() => _ChatPageScreenState();
}

class _ChatPageScreenState extends State<ChatPageScreen> {
  final ScrollController _scrollController = ScrollController();
  final AppDatabase _db = Get.find<AppDatabase>();
  final ApiClient _api = ApiClient.to;
  final AuthController _auth = Get.find<AuthController>();

  StreamSubscription? _messagesSubscription;
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _listenMessages();
    _connectWebSocket();
  }

  @override
  void dispose() {
    _messagesSubscription?.cancel();
    _scrollController.dispose();
    super.dispose();
  }

  void _listenMessages() {
    _messagesSubscription = _db.watchMessages(widget.chatId).listen((messages) {
      if (mounted) {
        setState(() {
          _isLoading = false;
        });
        // Auto scroll to bottom when new messages arrive
        WidgetsBinding.instance.addPostFrameCallback((_) {
          _scrollToBottom();
        });
      }
    });
  }

  void _connectWebSocket() {
    final ws = WebSocketService.to;
    if (!ws.isConnected) {
      ws.connect();
    }
  }

  void _scrollToBottom() {
    if (_scrollController.hasClients) {
      _scrollController.animateTo(
        _scrollController.position.maxScrollExtent,
        duration: const Duration(milliseconds: 300),
        curve: Curves.easeOut,
      );
    }
  }

  Future<void> _handleSend(String text) async {
    if (text.trim().isEmpty) return;

    final currentUserId = _auth.currentUser.value?.userId ?? 0;
    final now = DateTime.now();

    // 1. Insert message with "sending" status into Drift (optimistic UI)
    await _db.insertMessage(MessagesCompanion(
      seqId: const Value(0), // Will be updated after server response
      chatId: Value(widget.chatId),
      senderId: Value(currentUserId),
      type: const Value(1), // text
      content: Value(text),
      status: const Value('sending'),
      createdAt: Value(now),
    ));

    // Update chat session's last message
    await _updateChatSessionLastMessage(text, now);

    try {
      // 2. Call API to send message
      final response = await _api.post(
        '/messages',
        data: {
          'chat_id': widget.chatId,
          'type': 1, // text
          'content': text,
        },
      );

      final responseData = response.data;
      if (responseData is Map<String, dynamic>) {
        Map<String, dynamic> data;
        if (responseData.containsKey('data')) {
          data = responseData['data'] ?? {};
        } else {
          data = responseData;
        }

        // 3. Update message status to "sent" and update seqId
        // Find the message we just inserted (by content and createdAt)
        final messages = await _db.getMessagesByChatId(widget.chatId);
        final localMessage = messages.firstWhere(
          (m) => m.content == text && m.status == 'sending',
          orElse: () => messages.first,
        );

        final serverSeqId = data['seq_id'] as int? ?? 0;
        await _db.updateMessageStatus(localMessage.id, 'sent');

        // Update seqId if provided
        if (serverSeqId > 0) {
          await (_db.update(_db.messages)..where((t) => t.id.equals(localMessage.id)))
              .write(MessagesCompanion(seqId: Value(serverSeqId)));
        }
      } else {
        // API returned success but unexpected format
        await _db.updateMessageStatusByContent(widget.chatId, text, 'sending', 'sent');
      }
    } on DioException {
      // 4. Mark as failed on error
      await _db.updateMessageStatusByContent(widget.chatId, text, 'sending', 'failed');
    } catch (_) {
      await _db.updateMessageStatusByContent(widget.chatId, text, 'sending', 'failed');
    }
  }

  Future<void> _updateChatSessionLastMessage(String message, DateTime time) async {
    // Check if chat session exists
    final sessions = await (_db.select(_db.chatSessions)
          ..where((t) => t.chatId.equals(widget.chatId)))
        .get();

    if (sessions.isNotEmpty) {
      // Update existing session
      await (_db.update(_db.chatSessions)
            ..where((t) => t.chatId.equals(widget.chatId)))
          .write(ChatSessionsCompanion(
            lastMessage: Value(message),
            updatedAt: Value(time),
          ));
    } else {
      // Create new session
      await _db.insertOrUpdateChat(ChatSessionsCompanion(
        chatId: Value(widget.chatId),
        name: Value(widget.chatName),
        lastMessage: Value(message),
        unreadCount: const Value(0),
        updatedAt: Value(time),
      ));
    }
  }

  String _formatTime(DateTime time) {
    final hour = time.hour > 12 ? time.hour - 12 : time.hour;
    final period = time.hour >= 12 ? 'PM' : 'AM';
    final minute = time.minute.toString().padLeft(2, '0');
    return '$hour:$minute $period';
  }

  /// Convert Drift status string to MessageStatus enum
  MessageStatus? _convertStatus(String? status) {
    if (status == null) return null;
    switch (status) {
      case 'sending':
        return MessageStatus.sending;
      case 'sent':
        return MessageStatus.sent;
      case 'failed':
        return MessageStatus.failed;
      default:
        return null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final currentUserId = _auth.currentUser.value?.userId ?? 0;

    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
        title: Row(
          children: [
            CircleAvatar(
              radius: 18,
              backgroundColor: colorScheme.primaryContainer,
              child: Text(
                widget.avatarText,
                style: TextStyle(
                  color: colorScheme.onPrimaryContainer,
                  fontWeight: FontWeight.w600,
                  fontSize: 12,
                ),
              ),
            ),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    widget.chatName,
                    style: const TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                  const Text(
                    'Online',
                    style: TextStyle(
                      fontSize: 12,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: const Icon(Icons.more_vert),
            onPressed: () {
              // TODO: Show chat options
            },
          ),
        ],
      ),
      body: Column(
        children: [
          // Messages area with background
          Expanded(
            child: Container(
              color: colorScheme.surface,
              child: _isLoading
                  ? const Center(child: CircularProgressIndicator())
                  : StreamBuilder<List<Message>>(
                      stream: _db.watchMessages(widget.chatId),
                      builder: (context, snapshot) {
                        final messages = snapshot.data ?? [];
                        if (messages.isEmpty) {
                          return _buildEmptyState();
                        }
                        return _buildMessagesList(messages, currentUserId);
                      },
                    ),
            ),
          ),
          // Input bar
          Container(
            decoration: BoxDecoration(
              color: colorScheme.surface,
              boxShadow: [
                BoxShadow(
                  color: colorScheme.shadow.withValues(alpha: 0.05),
                  blurRadius: 10,
                  offset: const Offset(0, -2),
                ),
              ],
            ),
            child: SafeArea(
              top: false,
              child: ChatInputBar(
                onSend: _handleSend,
                onAttach: () {
                  // TODO: Show attachment options
                },
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.chat_bubble_outline,
            size: 64,
            color: Theme.of(context).colorScheme.outline,
          ),
          const SizedBox(height: 16),
          Text(
            'No messages yet',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: 8),
          Text(
            'Send a message to start the conversation',
            style: Theme.of(context).textTheme.bodyMedium,
          ),
        ],
      ),
    );
  }

  Widget _buildMessagesList(List<Message> messages, int currentUserId) {
    return Stack(
      children: [
        // Background pattern
        Positioned.fill(
          child: Opacity(
            opacity: 0.05,
            child: Container(
              decoration: BoxDecoration(
                color: Theme.of(context).colorScheme.primary,
              ),
            ),
          ),
        ),
        // Messages list
        ListView.builder(
          controller: _scrollController,
          padding: const EdgeInsets.only(top: 8, bottom: 8),
          itemCount: messages.length,
          itemBuilder: (context, index) {
            final message = messages[index];
            final isMe = message.senderId == currentUserId;

            return MessageBubble(
              content: message.content ?? '',
              isMe: isMe,
              isMarkdown: false,
              time: _formatTime(message.createdAt),
              status: _convertStatus(message.status),
            )
                .animate()
                .fadeIn(
                  duration: const Duration(milliseconds: 200),
                )
                .slideX(
                  begin: isMe ? 0.1 : -0.1,
                  end: 0,
                  duration: const Duration(milliseconds: 200),
                  curve: Curves.easeOut,
                );
          },
        ),
      ],
    );
  }
}
