import 'package:flutter/material.dart';
import 'package:flutter_animate/flutter_animate.dart';
import 'package:get/get.dart';

import '../controllers/contacts_controller.dart';
import '../../chat/screens/chat_page_screen.dart';

/// Contacts Screen - shows list of contacts
class ContactsScreen extends StatelessWidget {
  const ContactsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final controller = Get.find<ContactsController>();
    final colorScheme = Theme.of(context).colorScheme;

    return Scaffold(
      appBar: AppBar(
        title: const Text('Contacts'),
        actions: [
          IconButton(
            icon: const Icon(Icons.person_add_outlined),
            onPressed: () => _showAddContactDialog(context, controller),
          ),
          IconButton(
            icon: const Icon(Icons.sync),
            onPressed: () => controller.syncContacts(),
          ),
        ],
      ),
      body: Obx(() {
        if (controller.isLoading.value && controller.contacts.isEmpty) {
          return const Center(child: CircularProgressIndicator());
        }

        if (controller.contacts.isEmpty) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  Icons.contacts_outlined,
                  size: 64,
                  color: colorScheme.outline,
                ),
                const SizedBox(height: 16),
                Text(
                  'No contacts yet',
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: 8),
                Text(
                  'Add contacts to start chatting',
                  style: Theme.of(context).textTheme.bodyMedium,
                ),
                const SizedBox(height: 16),
                FilledButton.icon(
                  onPressed: () => _showAddContactDialog(context, controller),
                  icon: const Icon(Icons.person_add),
                  label: const Text('Add Contact'),
                ),
              ],
            ),
          );
        }

        return RefreshIndicator(
          onRefresh: () => controller.loadContacts(),
          child: ListView.builder(
            padding: const EdgeInsets.only(top: 8, bottom: 80),
            itemCount: controller.contacts.length,
            itemBuilder: (context, index) {
              final contact = controller.contacts[index];
              return _ContactTile(
                avatarText: contact.avatarText,
                name: contact.displayName,
                subtitle: contact.phone ?? '@${contact.username}',
                onTap: () {
                  // Navigate to chat
                  Navigator.of(context).push(
                    MaterialPageRoute(
                      builder: (context) => ChatPageScreen(
                        chatId: contact.id,
                        chatName: contact.displayName,
                        avatarText: contact.avatarText,
                      ),
                    ),
                  );
                },
                onLongPress: () => _showContactOptions(context, controller, contact),
              )
                  .animate()
                  .fadeIn(
                    delay: Duration(milliseconds: index * 30),
                    duration: const Duration(milliseconds: 200),
                  )
                  .slideX(
                    begin: 0.1,
                    end: 0,
                    delay: Duration(milliseconds: index * 30),
                    duration: const Duration(milliseconds: 200),
                  );
            },
          ),
        );
      }),
      floatingActionButton: FloatingActionButton(
        onPressed: () => _showAddContactDialog(context, controller),
        child: const Icon(Icons.person_add),
      ),
    );
  }

  void _showAddContactDialog(BuildContext context, ContactsController controller) {
    final textController = TextEditingController();

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('Add Contact'),
        content: TextField(
          controller: textController,
          decoration: const InputDecoration(
            labelText: 'Username',
            hintText: 'Enter username to add',
            prefixIcon: Icon(Icons.person_outline),
          ),
          autofocus: true,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('Cancel'),
          ),
          FilledButton(
            onPressed: () async {
              final username = textController.text.trim();
              if (username.isNotEmpty) {
                final success = await controller.addContact(username);
                if (success && context.mounted) {
                  Navigator.pop(context);
                  Get.snackbar(
                    'Success',
                    'Contact added successfully',
                    snackPosition: SnackPosition.BOTTOM,
                  );
                }
              }
            },
            child: const Text('Add'),
          ),
        ],
      ),
    );
  }

  void _showContactOptions(BuildContext context, ContactsController controller, Contact contact) {
    showModalBottomSheet(
      context: context,
      builder: (context) => SafeArea(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            ListTile(
              leading: const Icon(Icons.chat),
              title: const Text('Start Chat'),
              onTap: () {
                Navigator.pop(context);
                Navigator.of(context).push(
                  MaterialPageRoute(
                    builder: (context) => ChatPageScreen(
                      chatId: contact.id,
                      chatName: contact.displayName,
                      avatarText: contact.avatarText,
                    ),
                  ),
                );
              },
            ),
            ListTile(
              leading: Icon(Icons.delete_outline, color: Theme.of(context).colorScheme.error),
              title: Text('Delete Contact', style: TextStyle(color: Theme.of(context).colorScheme.error)),
              onTap: () async {
                Navigator.pop(context);
                final confirmed = await showDialog<bool>(
                  context: context,
                  builder: (context) => AlertDialog(
                    title: const Text('Delete Contact'),
                    content: Text('Are you sure you want to delete ${contact.displayName}?'),
                    actions: [
                      TextButton(
                        onPressed: () => Navigator.pop(context, false),
                        child: const Text('Cancel'),
                      ),
                      FilledButton(
                        onPressed: () => Navigator.pop(context, true),
                        style: FilledButton.styleFrom(
                          backgroundColor: Theme.of(context).colorScheme.error,
                        ),
                        child: const Text('Delete'),
                      ),
                    ],
                  ),
                );

                if (confirmed == true) {
                  await controller.deleteContact(contact.id);
                }
              },
            ),
          ],
        ),
      ),
    );
  }
}

class _ContactTile extends StatelessWidget {
  final String avatarText;
  final String name;
  final String subtitle;
  final VoidCallback onTap;
  final VoidCallback? onLongPress;

  const _ContactTile({
    required this.avatarText,
    required this.name,
    required this.subtitle,
    required this.onTap,
    this.onLongPress,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return ListTile(
      leading: CircleAvatar(
        backgroundColor: colorScheme.primaryContainer,
        child: Text(
          avatarText,
          style: TextStyle(
            color: colorScheme.onPrimaryContainer,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
      title: Text(name),
      subtitle: Text(
        subtitle,
        style: TextStyle(color: colorScheme.onSurfaceVariant),
      ),
      onTap: onTap,
      onLongPress: onLongPress,
    );
  }
}
