import 'package:drift/drift.dart';

import 'tables.dart';
import 'connection/connection.dart' as connection;

part 'database.g.dart';

@DriftDatabase(tables: [ChatSessions, Messages])
class AppDatabase extends _$AppDatabase {
  AppDatabase() : super(connection.connect());

  @override
  int get schemaVersion => 1;

  // ============ Watch Queries (Stream) ============

  /// Watch all chat sessions, ordered by updatedAt descending
  Stream<List<ChatSession>> watchAllChats() {
    return (select(chatSessions)
          ..orderBy([
            (t) => OrderingTerm(expression: t.updatedAt, mode: OrderingMode.desc)
          ]))
        .watch();
  }

  /// Watch messages for a specific chat, ordered by createdAt descending
  Stream<List<Message>> watchMessages(int chatId) {
    return (select(messages)
          ..where((t) => t.chatId.equals(chatId))
          ..orderBy([
            (t) => OrderingTerm(expression: t.createdAt, mode: OrderingMode.desc)
          ]))
        .watch();
  }

  // ============ Write Operations ============

  /// Insert or update a chat session
  Future<void> insertOrUpdateChat(ChatSessionsCompanion chat) async {
    await into(chatSessions).insertOnConflictUpdate(chat);
  }

  /// Insert a message
  Future<void> insertMessage(MessagesCompanion msg) async {
    await into(messages).insert(msg);
  }

  /// Update message status by id
  Future<void> updateMessageStatus(int id, String status) async {
    await (update(messages)..where((t) => t.id.equals(id)))
        .write(MessagesCompanion(status: Value(status)));
  }

  /// Update message status by content and current status (for optimistic UI updates)
  Future<void> updateMessageStatusByContent(
    int chatId,
    String content,
    String currentStatus,
    String newStatus,
  ) async {
    await (update(messages)
          ..where(
            (t) =>
                t.chatId.equals(chatId) &
                t.content.equals(content) &
                t.status.equals(currentStatus),
          ))
        .write(MessagesCompanion(status: Value(newStatus)));
  }

  /// Get messages by chatId with pagination
  Future<List<Message>> getMessagesByChatId(
    int chatId, {
    int limit = 50,
    int offset = 0,
  }) async {
    return (select(messages)
          ..where((t) => t.chatId.equals(chatId))
          ..orderBy([
            (t) => OrderingTerm(expression: t.createdAt, mode: OrderingMode.desc)
          ])
          ..limit(limit, offset: offset))
        .get();
  }

  /// Get max seqId from messages
  Future<int> getMaxSeqId() async {
    final result = await customSelect(
      'SELECT MAX(seq_id) as max_seq FROM messages',
    ).getSingleOrNull();
    return result?.read<int?>('max_seq') ?? 0;
  }

  /// Get messages with seqId greater than given value
  Future<List<Message>> getMessagesBySeqIdGreaterThan(
    int chatId,
    int seqId, {
    int limit = 50,
  }) async {
    return (select(messages)
          ..where((t) => t.chatId.equals(chatId) & t.seqId.isBiggerThanValue(seqId))
          ..orderBy([
            (t) => OrderingTerm(expression: t.seqId, mode: OrderingMode.asc)
          ])
          ..limit(limit))
        .get();
  }

  /// Delete a message by id
  Future<void> deleteMessage(int id) async {
    await (delete(messages)..where((t) => t.id.equals(id))).go();
  }

  /// Clear all data
  Future<void> clearAll() async {
    await delete(messages).go();
    await delete(chatSessions).go();
  }
}
