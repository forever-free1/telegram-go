/// Message send status
enum MessageStatus {
  sending,  // Sending to server
  sent,     // Successfully sent
  failed,   // Failed to send
}

/// Message model for local storage
class MessageModel {
  int id;
  String? localId; // For local temporary messages
  int seqId;
  int chatId;
  int senderId;
  int type;
  String? content;
  String? mediaUrl;
  int? duration;
  double? latitude;
  double? longitude;
  int? replyId;
  bool isRead;
  bool isDeleted;
  DateTime? readAt;
  DateTime createdAt;
  MessageStatus status; // Local status tracking

  MessageModel({
    this.id = 0,
    this.localId,
    this.seqId = 0,
    required this.chatId,
    required this.senderId,
    required this.type,
    this.content,
    this.mediaUrl,
    this.duration,
    this.latitude,
    this.longitude,
    this.replyId,
    this.isRead = false,
    this.isDeleted = false,
    this.readAt,
    DateTime? createdAt,
    this.status = MessageStatus.sent,
  }) : createdAt = createdAt ?? DateTime.now();

  Map<String, dynamic> toJson() => {
        'id': id,
        'local_id': localId,
        'seq_id': seqId,
        'chat_id': chatId,
        'sender_id': senderId,
        'type': type,
        'content': content,
        'media_url': mediaUrl,
        'duration': duration,
        'latitude': latitude,
        'longitude': longitude,
        'reply_id': replyId,
        'is_read': isRead,
        'is_deleted': isDeleted,
        'read_at': readAt?.toIso8601String(),
        'created_at': createdAt.toIso8601String(),
        'status': status.name,
      };

  factory MessageModel.fromJson(Map<String, dynamic> json) => MessageModel(
        id: json['id'] ?? 0,
        localId: json['local_id'],
        seqId: json['seq_id'] ?? 0,
        chatId: json['chat_id'] ?? 0,
        senderId: json['sender_id'] ?? 0,
        type: json['type'] ?? 1,
        content: json['content'],
        mediaUrl: json['media_url'],
        duration: json['duration'],
        latitude: json['latitude']?.toDouble(),
        longitude: json['longitude']?.toDouble(),
        replyId: json['reply_id'],
        isRead: json['is_read'] ?? false,
        isDeleted: json['is_deleted'] ?? false,
        readAt: json['read_at'] != null ? DateTime.tryParse(json['read_at']) : null,
        createdAt: json['created_at'] != null
            ? DateTime.tryParse(json['created_at']) ?? DateTime.now()
            : DateTime.now(),
        status: MessageStatus.values.firstWhere(
          (e) => e.name == json['status'],
          orElse: () => MessageStatus.sent,
        ),
      );

  static MessageModel fromServerJson(Map<String, dynamic> json) => MessageModel(
        id: json['id'] ?? 0,
        seqId: json['seq_id'] ?? 0,
        chatId: json['chat_id'] ?? 0,
        senderId: json['sender_id'] ?? 0,
        type: json['type'] ?? 1,
        content: json['content'],
        mediaUrl: json['media_url'],
        duration: json['duration'],
        latitude: json['latitude']?.toDouble(),
        longitude: json['longitude']?.toDouble(),
        replyId: json['reply_id'],
        isRead: json['is_read'] ?? false,
        isDeleted: json['is_deleted'] ?? false,
        readAt: json['read_at'] != null ? DateTime.tryParse(json['read_at']) : null,
        createdAt: json['created_at'] != null
            ? DateTime.tryParse(json['created_at']) ?? DateTime.now()
            : DateTime.now(),
        status: MessageStatus.sent,
      );

  /// Create a local temporary message (for optimistic UI)
  factory MessageModel.createLocal({
    required String localId,
    required int chatId,
    required int senderId,
    required String content,
  }) =>
      MessageModel(
        localId: localId,
        chatId: chatId,
        senderId: senderId,
        type: 1, // text
        content: content,
        status: MessageStatus.sending,
      );
}
