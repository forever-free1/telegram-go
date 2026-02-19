/// Chat session model for local storage
class ChatSessionModel {
  int id;
  int chatId;
  String name;
  String? avatarUrl;
  String? avatarText;
  String? lastMessage;
  DateTime? lastMessageTime;
  int unreadCount;
  DateTime updatedAt;

  ChatSessionModel({
    this.id = 0,
    required this.chatId,
    required this.name,
    this.avatarUrl,
    this.avatarText,
    this.lastMessage,
    this.lastMessageTime,
    this.unreadCount = 0,
    DateTime? updatedAt,
  }) : updatedAt = updatedAt ?? DateTime.now();

  Map<String, dynamic> toJson() => {
        'id': id,
        'chat_id': chatId,
        'name': name,
        'avatar_url': avatarUrl,
        'avatar_text': avatarText,
        'last_message': lastMessage,
        'last_message_time': lastMessageTime?.toIso8601String(),
        'unread_count': unreadCount,
        'updated_at': updatedAt.toIso8601String(),
      };

  factory ChatSessionModel.fromJson(Map<String, dynamic> json) => ChatSessionModel(
        id: json['id'] ?? 0,
        chatId: json['chat_id'] ?? 0,
        name: json['name'] ?? 'Unknown',
        avatarUrl: json['avatar_url'],
        avatarText: json['avatar_text'],
        lastMessage: json['last_message'],
        lastMessageTime: json['last_message_time'] != null
            ? DateTime.tryParse(json['last_message_time'])
            : null,
        unreadCount: json['unread_count'] ?? 0,
        updatedAt: json['updated_at'] != null
            ? DateTime.tryParse(json['updated_at']) ?? DateTime.now()
            : DateTime.now(),
      );

  static ChatSessionModel fromServerJson(Map<String, dynamic> json) => ChatSessionModel(
        chatId: json['id'] ?? 0,
        name: json['name'] ?? 'Unknown',
        avatarUrl: json['avatar_url'],
        avatarText: json['avatar_text'],
        lastMessage: json['last_message'],
        lastMessageTime: json['last_message_time'] != null
            ? DateTime.tryParse(json['last_message_time'])
            : null,
        unreadCount: json['unread_count'] ?? 0,
        updatedAt: json['updated_at'] != null
            ? DateTime.tryParse(json['updated_at']) ?? DateTime.now()
            : DateTime.now(),
      );

  void updateFromMessage(String message, DateTime time) {
    lastMessage = message;
    lastMessageTime = time;
    updatedAt = time;
  }
}
