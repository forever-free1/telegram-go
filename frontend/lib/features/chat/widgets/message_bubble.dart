import 'package:flutter/material.dart';
import 'package:flutter_markdown/flutter_markdown.dart';

import '../../../core/database/models/message_model.dart';

/// Message bubble widget - Kelivo style
class MessageBubble extends StatelessWidget {
  final String content;
  final bool isMe;
  final bool isMarkdown;
  final String time;
  final MessageStatus? status;

  const MessageBubble({
    super.key,
    required this.content,
    required this.isMe,
    this.isMarkdown = false,
    required this.time,
    this.status,
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
            // Time and status
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 4),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    time,
                    style: TextStyle(
                      color: colorScheme.outline,
                      fontSize: 11,
                    ),
                  ),
                  if (isMe && status != null) ...[
                    const SizedBox(width: 4),
                    _buildStatusIcon(status!, colorScheme),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusIcon(MessageStatus status, ColorScheme colorScheme) {
    switch (status) {
      case MessageStatus.sending:
        return SizedBox(
          width: 12,
          height: 12,
          child: CircularProgressIndicator(
            strokeWidth: 1.5,
            color: colorScheme.outline,
          ),
        );
      case MessageStatus.sent:
        return Icon(
          Icons.check,
          size: 12,
          color: colorScheme.outline,
        );
      case MessageStatus.failed:
        return Icon(
          Icons.error_outline,
          size: 12,
          color: colorScheme.error,
        );
    }
  }
}
