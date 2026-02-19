import 'package:get/get.dart';
import 'package:dio/dio.dart';

import '../../../core/network/api_client.dart';

/// Contact model
class Contact {
  final int id;
  final String username;
  final String? nickname;
  final String? avatar;
  final String? phone;
  final String? bio;
  final int status;

  Contact({
    required this.id,
    required this.username,
    this.nickname,
    this.avatar,
    this.phone,
    this.bio,
    this.status = 1,
  });

  String get displayName => nickname?.isNotEmpty == true ? nickname! : username;

  String get avatarText {
    if (displayName.isEmpty) return '?';
    final parts = displayName.trim().split(' ');
    if (parts.length >= 2) {
      return '${parts[0][0]}${parts[1][0]}'.toUpperCase();
    }
    return displayName[0].toUpperCase();
  }

  factory Contact.fromJson(Map<String, dynamic> json) {
    return Contact(
      id: json['id'] ?? 0,
      username: json['username'] ?? '',
      nickname: json['nickname'],
      avatar: json['avatar'],
      phone: json['phone'],
      bio: json['bio'],
      status: json['status'] ?? 1,
    );
  }
}

/// Contacts Controller
class ContactsController extends GetxController {
  final ApiClient _api = ApiClient.to;

  final RxBool isLoading = false.obs;
  final RxList<Contact> contacts = <Contact>[].obs;
  final RxString errorMessage = ''.obs;

  @override
  void onInit() {
    super.onInit();
    loadContacts();
  }

  /// Load contacts from API
  Future<void> loadContacts() async {
    isLoading.value = true;
    errorMessage.value = '';

    try {
      final response = await _api.get('/contacts');
      final responseData = response.data;

      List<dynamic> contactsList = [];
      if (responseData is Map<String, dynamic>) {
        if (responseData.containsKey('data')) {
          contactsList = responseData['data'] as List<dynamic>? ?? [];
        } else {
          contactsList = responseData['contacts'] as List<dynamic>? ?? [];
        }
      }

      contacts.value = contactsList
          .map((json) => Contact.fromJson(json as Map<String, dynamic>))
          .toList();
    } on DioException catch (e) {
      if (e.response?.statusCode == 401) {
        errorMessage.value = 'Session expired';
      } else {
        errorMessage.value = 'Failed to load contacts';
      }
    } catch (e) {
      errorMessage.value = 'Failed to load contacts';
    } finally {
      isLoading.value = false;
    }
  }

  /// Add a new contact
  Future<bool> addContact(String username) async {
    isLoading.value = true;
    errorMessage.value = '';

    try {
      await _api.post(
        '/contacts',
        data: {'username': username},
      );

      // Reload contacts after adding
      await loadContacts();
      return true;
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) {
        errorMessage.value = 'User not found';
      } else if (e.response?.statusCode == 400) {
        errorMessage.value = 'Cannot add yourself as contact';
      } else {
        errorMessage.value = 'Failed to add contact';
      }
      return false;
    } catch (e) {
      errorMessage.value = 'Failed to add contact';
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /// Delete a contact
  Future<bool> deleteContact(int contactId) async {
    isLoading.value = true;
    errorMessage.value = '';

    try {
      await _api.delete('/contacts/$contactId');

      // Remove from local list
      contacts.removeWhere((c) => c.id == contactId);
      return true;
    } on DioException {
      errorMessage.value = 'Failed to delete contact';
      return false;
    } catch (_) {
      errorMessage.value = 'Failed to delete contact';
      return false;
    } finally {
      isLoading.value = false;
    }
  }

  /// Sync contacts from server
  Future<void> syncContacts() async {
    isLoading.value = true;
    errorMessage.value = '';

    try {
      // Get local max contact ID (for incremental sync)
      final maxId = contacts.isEmpty ? 0 : contacts.map((c) => c.id).reduce((a, b) => a > b ? a : b);

      final response = await _api.post(
        '/contacts/sync',
        data: {'last_contact_id': maxId},
      );

      final responseData = response.data;
      if (responseData is Map<String, dynamic> && responseData.containsKey('data')) {
        final data = responseData['data'] as Map<String, dynamic>;
        final newContacts = data['contacts'] as List<dynamic>? ?? [];

        for (final json in newContacts) {
          final contact = Contact.fromJson(json as Map<String, dynamic>);
          // Update or add
          final index = contacts.indexWhere((c) => c.id == contact.id);
          if (index >= 0) {
            contacts[index] = contact;
          } else {
            contacts.add(contact);
          }
        }
      }
    } catch (e) {
      // Silently fail for sync
    } finally {
      isLoading.value = false;
    }
  }
}
