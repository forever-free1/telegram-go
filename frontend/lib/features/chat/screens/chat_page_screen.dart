import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../widgets/message_bubble.dart';
import '../widgets/chat_input_bar.dart';

/// Mock message data
class MessageData {
  final int id;
  final String content;
  final bool isMe;
  final bool isMarkdown;
  final String time;

  MessageData({
    required this.id,
    required this.content,
    required this.isMe,
    this.isMarkdown = false,
    required this.time,
  });
}

final List<MessageData> mockMessages = [
  MessageData(id: 1, content: 'Hey! How are you doing?', isMe: false, time: '10:30 AM'),
  MessageData(id: 2, content: 'I\'m good, thanks! Just working on a new Flutter project.', isMe: true, time: '10:31 AM'),
  MessageData(id: 3, content: 'That sounds cool! What kind of app are you building?', isMe: false, time: '10:32 AM'),
  MessageData(id: 4, content: 'A Telegram clone with Material Design 3. Check out this code:', isMe: true, time: '10:33 AM'),
  MessageData(
    id: 5,
    content: '''
```dart
void main() {
  print('Hello, World!');
}
```

It's pretty slick!''', isMe: true, isMarkdown: true, time: '10:33 AM'),
  MessageData(id: 6, content: 'Wow, that looks amazing! 😍', isMe: false, time: '10:34 AM'),
  MessageData(id: 7, content: 'Yeah, I\'m using flutter_animate for smooth animations.', isMe: true, time: '10:35 AM'),
  MessageData(id: 8, content: 'You should check out Kelivo for UI inspiration. Their design is really clean.', isMe: false, time: '10:36 AM'),
  MessageData(id: 9, content: 'Already did! The chat bubbles and input bar are inspired by their design.', isMe: true, time: '10:37 AM'),
  MessageData(id: 10, content: 'Nice! Let me know when it\'s ready to test 👀', isMe: false, time: '10:38 AM'),
];

/// Chat page screen
class ChatPageScreen extends StatefulWidget {
  final String chatName;
  final String avatarText;

  const ChatPageScreen({
    super.key,
    required this.chatName,
    required this.avatarText,
  });

  @override
  State<ChatPageScreen> createState() => _ChatPageScreenState();
}

class _ChatPageScreenState extends State<ChatPageScreen> {
  final ScrollController _scrollController = ScrollController();
  late List<MessageData> _messages;

  @override
  void initState() {
    super.initState();
    _messages = List.from(mockMessages);
    // Scroll to bottom after build
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _scrollToBottom();
    });
  }

  @override
  void dispose() {
    _scrollController.dispose();
    super.dispose();
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

  void _handleSend(String text) {
    setState(() {
      _messages.add(MessageData(
        id: _messages.length + 1,
        content: text,
        isMe: true,
        time: _formatTime(DateTime.now()),
      ));
    });
    _scrollToBottom();
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
              child: Stack(
                children: [
                  // Background pattern (for future expansion)
                  Positioned.fill(
                    child: Opacity(
                      opacity: 0.05,
                      child: Container(
                        decoration: BoxDecoration(
                          // Placeholder for background pattern
                          color: colorScheme.primary,
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
                      return MessageBubble(
                        content: message.content,
                        isMe: message.isMe,
                        isMarkdown: message.isMarkdown,
                        time: message.time,
                      )
                          .animate()
                          .fadeIn(
                            duration: const Duration(milliseconds: 200),
                          )
                          .slideX(
                            begin: message.isMe ? 0.1 : -0.1,
                            end: 0,
                            duration: const Duration(milliseconds: 200),
                            curve: Curves.easeOut,
                          );
                    },
                  ),
                ],
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
}
