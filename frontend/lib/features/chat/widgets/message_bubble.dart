import 'package:flutter/material.dart';
import 'package:flutter_markdown/flutter_markdown.dart';

/// Message bubble widget - Kelivo style
class MessageBubble extends StatelessWidget {
  final String content;
  final bool isMe;
  final bool isMarkdown;
  final String time;

  const MessageBubble({
    super.key,
    required this.content,
    required this.isMe,
    this.isMarkdown = false,
    required this.time,
  });

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;
    final screenWidth = MediaQuery.of(context).size.width;
    final maxWidth = screenWidth * 0.75;

    return Align(
      alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: BoxConstraints(maxWidth: maxWidth),
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        child: Column(
          crossAxisAlignment: isMe ? CrossAxisAlignment.end : CrossAxisAlignment.start,
          children: [
            // Message bubble
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
              decoration: BoxDecoration(
                color: isMe ? colorScheme.primary : colorScheme.secondaryContainer,
                borderRadius: BorderRadius.only(
                  topLeft: const Radius.circular(20),
                  topRight: const Radius.circular(20),
                  bottomLeft: isMe ? const Radius.circular(20) : const Radius.circular(4),
                  bottomRight: isMe ? const Radius.circular(4) : const Radius.circular(20),
                ),
              ),
              child: isMarkdown
                  ? MarkdownBody(
                      data: content,
                      styleSheet: MarkdownStyleSheet(
                        p: TextStyle(
                          color: isMe ? colorScheme.onPrimary : colorScheme.onSecondaryContainer,
                          fontSize: 15,
                        ),
                        code: TextStyle(
                          color: isMe ? colorScheme.onPrimary : colorScheme.onSecondaryContainer,
                          backgroundColor: isMe
                              ? colorScheme.onPrimary.withValues(alpha: 0.1)
                              : colorScheme.onSecondaryContainer.withValues(alpha: 0.1),
                          fontFamily: 'monospace',
                        ),
                        codeblockDecoration: BoxDecoration(
                          color: isMe
                              ? colorScheme.onPrimary.withValues(alpha: 0.1)
                              : colorScheme.onSecondaryContainer.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        blockquoteDecoration: BoxDecoration(
                          color: isMe
                              ? colorScheme.onPrimary.withValues(alpha: 0.1)
                              : colorScheme.onSecondaryContainer.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(8),
                        ),
                      ),
                    )
                  : Text(
                      content,
                      style: TextStyle(
                        color: isMe ? colorScheme.onPrimary : colorScheme.onSecondaryContainer,
                        fontSize: 15,
                      ),
                    ),
            ),
            const SizedBox(height: 2),
            // Time
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Text(
                time,
                style: TextStyle(
                  color: colorScheme.outline,
                  fontSize: 11,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
