import 'dart:async';
import 'dart:convert';
import 'package:get/get.dart';
import 'package:shared_preferences/shared_preferences.dart';

import 'models/message_model.dart';
import 'models/chat_session_model.dart';

/// Database service - SharedPreferences-based storage
class DatabaseService extends GetxService {
  static DatabaseService get to => Get.find<DatabaseService>();

  static const String _messagesKey = 'db_messages';
  static const String _chatSessionsKey = 'db_chat_sessions';
  static const String _maxSeqIdKey = 'db_max_seq_id';

  late SharedPreferences _prefs;
  final _messagesController = StreamController<List<MessageModel>>.broadcast();
  final _chatSessionsController = StreamController<List<ChatSessionModel>>.broadcast();

  Stream<List<MessageModel>> get messagesStream => _messagesController.stream;
  Stream<List<ChatSessionModel>> get chatSessionsStream => _chatSessionsController.stream;

  Future<DatabaseService> init() async {
    _prefs = await SharedPreferences.getInstance();
    _notifyChatSessionsChanged();
    return this;
  }

  // Message operations
  Future<List<MessageModel>> getMessagesByChatId(int chatId, {int limit = 50, int offset = 0}) async {
    final messages = await _getAllMessages();
    final filtered = messages.where((m) => m.chatId == chatId && !m.isDeleted).toList();
    filtered.sort((a, b) => b.createdAt.compareTo(a.createdAt));

    final start = offset;
    final end = (start + limit).clamp(0, filtered.length);
    if (start >= filtered.length) return [];

    return filtered.sublist(start, end);
  }

  Future<int> getMaxSeqId() async {
    return _prefs.getInt(_maxSeqIdKey) ?? 0;
  }

  Future<void> saveMessages(List<MessageModel> messages) async {
    final allMessages = await _getAllMessages();

    for (final msg in messages) {
      // Check if message already exists by localId or seqId
      int existingIndex = -1;
      if (msg.localId != null) {
        existingIndex = allMessages.indexWhere((m) => m.localId == msg.localId);
      }
      if (existingIndex < 0 && msg.seqId > 0) {
        existingIndex = allMessages.indexWhere((m) => m.seqId == msg.seqId);
      }

      if (existingIndex >= 0) {
        allMessages[existingIndex] = msg;
      } else {
        allMessages.add(msg);
      }

      // Update max seq ID
      if (msg.seqId > 0) {
        final currentMax = await getMaxSeqId();
        if (msg.seqId > currentMax) {
          await _prefs.setInt(_maxSeqIdKey, msg.seqId);
        }
      }
    }

    await _saveAllMessages(allMessages);
  }

  /// Save a single message
  Future<void> saveMessage(MessageModel message) async {
    await saveMessages([message]);
  }

  /// Update message status by localId
  Future<void> updateMessageStatus(String localId, MessageStatus status, {int? serverSeqId}) async {
    final allMessages = await _getAllMessages();
    final index = allMessages.indexWhere((m) => m.localId == localId);

    if (index >= 0) {
      allMessages[index].status = status;
      if (serverSeqId != null) {
        allMessages[index].seqId = serverSeqId;
      }
      await _saveAllMessages(allMessages);
    }
  }

  // Chat session operations
  Future<List<ChatSessionModel>> getAllChatSessions() async {
    final sessions = await _getAllChatSessions();
    sessions.sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
    return sessions;
  }

  Stream<List<ChatSessionModel>> watchChatSessions() {
    // Emit current data immediately
    _notifyChatSessionsChanged();
    return _chatSessionsController.stream;
  }

  Future<ChatSessionModel?> getChatSession(int chatId) async {
    final sessions = await _getAllChatSessions();
    try {
      return sessions.firstWhere((s) => s.chatId == chatId);
    } catch (_) {
      return null;
    }
  }

  Future<void> saveChatSession(ChatSessionModel session) async {
    final sessions = await _getAllChatSessions();

    final existingIndex = sessions.indexWhere((s) => s.chatId == session.chatId);
    if (existingIndex >= 0) {
      sessions[existingIndex] = session;
    } else {
      sessions.add(session);
    }

    await _saveAllChatSessions(sessions);
    _notifyChatSessionsChanged();
  }

  Future<void> updateChatSessionLastMessage(int chatId, String message, DateTime time) async {
    final session = await getChatSession(chatId);
    if (session != null) {
      session.updateFromMessage(message, time);
      await saveChatSession(session);
    }
  }

  Future<void> incrementUnreadCount(int chatId) async {
    final session = await getChatSession(chatId);
    if (session != null) {
      session.unreadCount++;
      await saveChatSession(session);
    }
  }

  Future<void> clearUnreadCount(int chatId) async {
    final session = await getChatSession(chatId);
    if (session != null) {
      session.unreadCount = 0;
      await saveChatSession(session);
    }
  }

  /// Clear all data (for logout)
  Future<void> clearAll() async {
    await _prefs.remove(_messagesKey);
    await _prefs.remove(_chatSessionsKey);
    await _prefs.remove(_maxSeqIdKey);
    _notifyChatSessionsChanged();
  }

  // Private helpers
  Future<List<MessageModel>> _getAllMessages() async {
    final jsonStr = _prefs.getString(_messagesKey);
    if (jsonStr == null || jsonStr.isEmpty) return [];

    final List<dynamic> jsonList = json.decode(jsonStr);
    return jsonList.map((e) => MessageModel.fromJson(e)).toList();
  }

  Future<void> _saveAllMessages(List<MessageModel> messages) async {
    final jsonStr = json.encode(messages.map((e) => e.toJson()).toList());
    await _prefs.setString(_messagesKey, jsonStr);
  }

  Future<List<ChatSessionModel>> _getAllChatSessions() async {
    final jsonStr = _prefs.getString(_chatSessionsKey);
    if (jsonStr == null || jsonStr.isEmpty) return [];

    final List<dynamic> jsonList = json.decode(jsonStr);
    return jsonList.map((e) => ChatSessionModel.fromJson(e)).toList();
  }

  Future<void> _saveAllChatSessions(List<ChatSessionModel> sessions) async {
    final jsonStr = json.encode(sessions.map((e) => e.toJson()).toList());
    await _prefs.setString(_chatSessionsKey, jsonStr);
  }

  void _notifyChatSessionsChanged() async {
    final sessions = await _getAllChatSessions();
    sessions.sort((a, b) => b.updatedAt.compareTo(a.updatedAt));
    _chatSessionsController.add(sessions);
  }

  void dispose() {
    _messagesController.close();
    _chatSessionsController.close();
  }
}
