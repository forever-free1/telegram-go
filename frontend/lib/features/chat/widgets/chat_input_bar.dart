import 'package:flutter/material.dart';

/// Chat input bar widget - Kelivo style floating input
class ChatInputBar extends StatefulWidget {
  final Function(String) onSend;
  final VoidCallback? onAttach;

  const ChatInputBar({
    super.key,
    required this.onSend,
    this.onAttach,
  });

  @override
  State<ChatInputBar> createState() => _ChatInputBarState();
}

class _ChatInputBarState extends State<ChatInputBar> {
  final TextEditingController _controller = TextEditingController();
  final FocusNode _focusNode = FocusNode();
  bool _hasText = false;

  @override
  void initState() {
    super.initState();
    _controller.addListener(() {
      final hasText = _controller.text.trim().isNotEmpty;
      if (hasText != _hasText) {
        setState(() {
          _hasText = hasText;
        });
      }
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    _focusNode.dispose();
    super.dispose();
  }

  void _handleSend() {
    final text = _controller.text.trim();
    if (text.isNotEmpty) {
      widget.onSend(text);
      _controller.clear();
    }
  }

  @override
  Widget build(BuildContext context) {
    final colorScheme = Theme.of(context).colorScheme;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      child: Row(
        children: [
          // Attachment button
          if (widget.onAttach != null)
            IconButton(
              onPressed: widget.onAttach,
              icon: Icon(
                Icons.add_circle_outline,
                color: colorScheme.onSurfaceVariant,
              ),
            ),
          // Input field
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                color: colorScheme.surfaceContainerHighest,
                borderRadius: BorderRadius.circular(24),
              ),
              child: Row(
                children: [
                  // Emoji button
                  IconButton(
                    onPressed: () {
                      // TODO: Implement emoji picker
                    },
                    icon: Icon(
                      Icons.emoji_emotions_outlined,
                      color: colorScheme.onSurfaceVariant,
                    ),
                  ),
                  // Text field
                  Expanded(
                    child: TextField(
                      controller: _controller,
                      focusNode: _focusNode,
                      decoration: InputDecoration(
                        hintText: 'Message...',
                        hintStyle: TextStyle(
                          color: colorScheme.onSurfaceVariant.withValues(alpha: 0.6),
                        ),
                        border: InputBorder.none,
                        contentPadding: const EdgeInsets.symmetric(vertical: 10),
                      ),
                      textCapitalization: TextCapitalization.sentences,
                      maxLines: 5,
                      minLines: 1,
                      onSubmitted: (_) => _handleSend(),
                    ),
                  ),
                  // Send button
                  if (_hasText)
                    IconButton(
                      onPressed: _handleSend,
                      icon: Icon(
                        Icons.send_rounded,
                        color: colorScheme.primary,
                      ),
                    )
                  else
                    IconButton(
                      onPressed: () {
                        // TODO: Implement voice message
                      },
                      icon: Icon(
                        Icons.mic_none,
                        color: colorScheme.onSurfaceVariant,
                      ),
                    ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }
}
