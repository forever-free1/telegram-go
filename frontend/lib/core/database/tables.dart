import 'package:drift/drift.dart';

/// Chat sessions table
class ChatSessions extends Table {
  IntColumn get chatId => integer()();
  TextColumn get name => text()();
  TextColumn get avatarUrl => text().nullable()();
  TextColumn get lastMessage => text().nullable()();
  IntColumn get unreadCount => integer().withDefault(const Constant(0))();
  DateTimeColumn get updatedAt => dateTime()();

  @override
  Set<Column> get primaryKey => {chatId};
}

/// Messages table
class Messages extends Table {
  IntColumn get id => integer().autoIncrement()();
  IntColumn get seqId => integer()();
  IntColumn get chatId => integer().references(ChatSessions, #chatId)();
  IntColumn get senderId => integer()();
  IntColumn get type => integer()();
  TextColumn get content => text().nullable()();
  TextColumn get status => text().withDefault(const Constant('sending'))();
  DateTimeColumn get createdAt => dateTime()();
}
