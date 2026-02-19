import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:get/get.dart';

import '../../../core/database/database_service.dart';
import '../../../core/database/models/chat_session_model.dart';
import '../../../core/sync/sync_controller.dart';
import '../widgets/chat_tile.dart';
import 'chat_page_screen.dart';

/// Chat list screen with Material 3 large title
class ChatListScreen extends StatelessWidget {
  const ChatListScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final db = DatabaseService.to;
    final syncController = Get.find<SyncController>();

    // Trigger initial sync if needed
    syncController.syncMessages();

    return Scaffold(
      body: CustomScrollView(
        slivers: [
          // Large title AppBar - Material 3 style
          SliverAppBar.large(
            title: const Text('Messages'),
            actions: [
              // Sync indicator
              Obx(() {
                if (syncController.isSyncing.value) {
                  return const Padding(
                    padding: EdgeInsets.all(12.0),
                    child: SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  );
                }
                return IconButton(
                  icon: const Icon(Icons.sync),
                  onPressed: () => syncController.syncMessages(),
                );
              }),
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
          // Chat list from Isar database
          StreamBuilder<List<ChatSessionModel>>(
            stream: db.watchChatSessions(),
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const SliverFillRemaining(
                  child: Center(
                    child: CircularProgressIndicator(),
                  ),
                );
              }

              final chats = snapshot.data ?? [];

              if (chats.isEmpty) {
                // Show empty state with hint to sync
                return SliverFillRemaining(
                  child: Center(
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
                          'No chats yet',
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        const SizedBox(height: 8),
                        Text(
                          'Start a new conversation or sync your chats',
                          style: Theme.of(context).textTheme.bodyMedium,
                          textAlign: TextAlign.center,
                        ),
                        const SizedBox(height: 16),
                        FilledButton.icon(
                          onPressed: () => syncController.syncMessages(),
                          icon: const Icon(Icons.sync),
                          label: const Text('Sync Now'),
                        ),
                      ],
                    ),
                  ),
                );
              }

              return SliverPadding(
                padding: const EdgeInsets.only(top: 8, bottom: 80),
                sliver: SliverList(
                  delegate: SliverChildBuilderDelegate(
                    (context, index) {
                      final chat = chats[index];
                      return ChatTile(
                        avatarText: chat.avatarText ?? _getInitials(chat.name),
                        title: chat.name,
                        subtitle: chat.lastMessage ?? 'No messages yet',
                        time: _formatTime(chat.lastMessageTime),
                        unreadCount: chat.unreadCount > 0 ? chat.unreadCount : null,
                        onTap: () {
                          Navigator.of(context).push(
                            MaterialPageRoute(
                              builder: (context) => ChatPageScreen(
                                chatId: chat.chatId,
                                chatName: chat.name,
                                avatarText: chat.avatarText ?? _getInitials(chat.name),
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
                    childCount: chats.length,
                  ),
                ),
              );
            },
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

  /// Get initials from name
  String _getInitials(String name) {
    final parts = name.trim().split(' ');
    if (parts.length >= 2) {
      return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
    }
    return name.isNotEmpty ? name[0].toUpperCase() : '?';
  }

  /// Format time for display
  String _formatTime(DateTime? time) {
    if (time == null) return '';

    final now = DateTime.now();
    final diff = now.difference(time);

    if (diff.inDays == 0) {
      return '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
    } else if (diff.inDays == 1) {
      return 'Yesterday';
    } else if (diff.inDays < 7) {
      const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
      return days[time.weekday - 1];
    } else {
      return '${time.month}/${time.day}';
    }
  }
}
