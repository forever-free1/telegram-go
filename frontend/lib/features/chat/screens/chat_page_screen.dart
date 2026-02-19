import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:get/get.dart';
import 'package:dio/dio.dart';

import '../../../core/database/database_service.dart';
import '../../../core/database/models/message_model.dart';
import '../../../core/network/api_client.dart';
import '../../../core/websocket/websocket_service.dart';
import '../../auth/controllers/auth_controller.dart';
import '../widgets/message_bubble.dart';
import '../widgets/chat_input_bar.dart';

/// Chat page screen
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
  final DatabaseService _db = DatabaseService.to;
  final ApiClient _api = ApiClient.to;
  final AuthController _auth = Get.find<AuthController>();

  List<MessageModel> _messages = [];
  bool _isLoading = true;

  @override
  void initState() {
    super.initState();
    _loadMessages();
    _connectWebSocket();
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
  }

  Future<void> _loadMessages() async {
    final messages = await _db.getMessagesByChatId(widget.chatId);
    setState(() {
      _messages = messages;
      _isLoading = false;
    });
    _scrollToBottom();
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

    // Generate local ID for tracking
    final localId = '${DateTime.now().millisecondsSinceEpoch}_${widget.chatId}';
    final currentUserId = _auth.currentUser.value?.userId ?? 0;

    // 1. Create local message with "sending" status
    final localMessage = MessageModel.createLocal(
      localId: localId,
      chatId: widget.chatId,
      senderId: currentUserId,
      content: text,
    );

    // Save to local DB immediately (optimistic UI)
    await _db.saveMessage(localMessage);
    await _db.updateChatSessionLastMessage(
      widget.chatId,
      text,
      localMessage.createdAt,
    );

    // Reload messages to update UI
    await _loadMessages();

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

        // 3. Update local message status to "sent"
        final serverMessage = MessageModel.fromServerJson(data);
        await _db.updateMessageStatus(localId, MessageStatus.sent, serverSeqId: serverMessage.seqId);
      } else {
        // API returned success but unexpected format
        await _db.updateMessageStatus(localId, MessageStatus.sent);
      }
    } on DioException {
      // 4. Mark as failed on error
      await _db.updateMessageStatus(localId, MessageStatus.failed);
    } catch (_) {
      await _db.updateMessageStatus(localId, MessageStatus.failed);
    }

    // Reload to reflect status change
    await _loadMessages();
  }

  String _formatTime(DateTime time) {
    final hour = time.hour > 12 ? time.hour - 12 : time.hour;
    final period = time.hour >= 12 ? 'PM' : 'AM';
    final minute = time.minute.toString().padLeft(2, '0');
    return '$hour:$minute $period';
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
                  Text(
                    'Online',
                    style: TextStyle(
                      fontSize: 12,
                      color: colorScheme.primary,
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
                  : _messages.isEmpty
                      ? _buildEmptyState()
                      : _buildMessagesList(currentUserId),
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

  Widget _buildMessagesList(int currentUserId) {
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
          itemCount: _messages.length,
          itemBuilder: (context, index) {
            final message = _messages[index];
            final isMe = message.senderId == currentUserId;

            return MessageBubble(
              content: message.content ?? '',
              isMe: isMe,
              isMarkdown: false,
              time: _formatTime(message.createdAt),
              status: message.status,
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
