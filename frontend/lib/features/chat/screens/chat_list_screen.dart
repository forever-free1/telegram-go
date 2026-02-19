import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../widgets/chat_tile.dart';
import 'chat_page_screen.dart';

/// Mock data for chat list
class ChatData {
  final int id;
  final String name;
  final String avatarText;
  final String lastMessage;
  final String time;
  final int? unreadCount;

  ChatData({
    required this.id,
    required this.name,
    required this.avatarText,
    required this.lastMessage,
    required this.time,
    this.unreadCount,
  });
}

final List<ChatData> mockChats = [
  ChatData(
    id: 1,
    name: 'Alice Johnson',
    avatarText: 'AJ',
    lastMessage: 'Hey, how are you doing?',
    time: '12:30',
    unreadCount: 3,
  ),
  ChatData(
    id: 2,
    name: 'Bob Smith',
    avatarText: 'BS',
    lastMessage: 'See you tomorrow!',
    time: '11:45',
  ),
  ChatData(
    id: 3,
    name: 'Carol White',
    avatarText: 'CW',
    lastMessage: 'Thanks for the help 🙏',
    time: 'Yesterday',
    unreadCount: 1,
  ),
  ChatData(
    id: 4,
    name: 'David Brown',
    avatarText: 'DB',
    lastMessage: 'The meeting is at 3pm',
    time: 'Yesterday',
  ),
  ChatData(
    id: 5,
    name: 'Eve Davis',
    avatarText: 'ED',
    lastMessage: 'Can you send me the file?',
    time: 'Mon',
    unreadCount: 12,
  ),
  ChatData(
    id: 6,
    name: 'Frank Miller',
    avatarText: 'FM',
    lastMessage: '👍',
    time: 'Mon',
  ),
  ChatData(
    id: 7,
    name: 'Grace Lee',
    avatarText: 'GL',
    lastMessage: 'Let me know when you\'re free',
    time: 'Sun',
  ),
  ChatData(
    id: 8,
    name: 'Henry Wilson',
    avatarText: 'HW',
    lastMessage: 'Great work on the project!',
    time: 'Sun',
  ),
];

/// Chat list screen with Material 3 large title
class ChatListScreen extends StatelessWidget {
  const ChatListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: CustomScrollView(
        slivers: [
          // Large title AppBar - Material 3 style
          SliverAppBar.large(
            title: const Text('Messages'),
            actions: [
              IconButton(
                icon: const Icon(Icons.search),
                onPressed: () {
                  // TODO: Implement search
                },
              ),
              IconButton(
                icon: const Icon(Icons.edit_square),
                onPressed: () {
                  // TODO: Implement new chat
                },
              ),
            ],
          ),
          // Chat list
          SliverPadding(
            padding: const EdgeInsets.only(top: 8, bottom: 80),
            sliver: SliverList(
              delegate: SliverChildBuilderDelegate(
                (context, index) {
                  final chat = mockChats[index];
                  return ChatTile(
                    avatarText: chat.avatarText,
                    title: chat.name,
                    subtitle: chat.lastMessage,
                    time: chat.time,
                    unreadCount: chat.unreadCount,
                    onTap: () {
                      Navigator.of(context).push(
                        MaterialPageRoute(
                          builder: (context) => ChatPageScreen(
                            chatName: chat.name,
                            avatarText: chat.avatarText,
                          ),
                        ),
                      );
                    },
                  )
                      .animate()
                      .fadeIn(
                        delay: Duration(milliseconds: index * 50),
                        duration: const Duration(milliseconds: 300),
                      )
                      .slideY(
                        begin: 0.2,
                        end: 0,
                        delay: Duration(milliseconds: index * 50),
                        duration: const Duration(milliseconds: 300),
                        curve: Curves.easeOut,
                      );
                },
                childCount: mockChats.length,
              ),
            ),
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          // TODO: Implement new chat
        },
        child: const Icon(Icons.add),
      ),
    );
  }
}
