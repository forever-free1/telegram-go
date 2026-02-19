import 'package:get/get.dart';
import 'package:dio/dio.dart';
import '../../../core/network/api_client.dart';
import '../../../features/chat/screens/chat_list_screen.dart';

/// Login Request DTO
class LoginRequest {
  final String username;
  final String password;

  LoginRequest({required this.username, required this.password});

  Map<String, dynamic> toJson() => {
        'username': username,
        'password': password,
      };
}

/// Login Response DTO
class LoginResponse {
  final String token;
  final int userId;
  final String? username;
  final String? nickname;

  LoginResponse({
    required this.token,
    required this.userId,
    this.username,
    this.nickname,
  });

  factory LoginResponse.fromJson(Map<String, dynamic> json) {
    return LoginResponse(
      token: json['token'] ?? '',
      userId: json['user_id'] ?? json['userId'] ?? 0,
      username: json['username'],
      nickname: json['nickname'],
    );
  }
}

/// Auth Controller - Handles authentication
class AuthController extends GetxController {
  final ApiClient _apiClient = ApiClient.to;

  final RxBool isLoading = false.obs;
  final RxString errorMessage = ''.obs;
  final Rx<LoginResponse?> currentUser = Rx<LoginResponse?>(null);

  @override
  void onInit() {
    super.onInit();
    checkLoginStatus();
  }

  /// Check if user is already logged in
  Future<void> checkLoginStatus() async {
    final hasToken = await ApiClient.hasToken();
    if (hasToken) {
      // Token exists, user is logged in
      // In a real app, you might want to validate the token with the server
      Get.offAll(() => const ChatListScreen());
    }
  }

  /// Handle unauthorized response - called by interceptor
  void handleUnauthorized() {
    errorMessage.value = 'Session expired. Please login again.';
    currentUser.value = null;
    Get.offAllNamed('/login');
  }

  /// Login with username and password
  Future<void> login(String username, String password) async {
    if (username.isEmpty || password.isEmpty) {
      errorMessage.value = 'Please enter username and password';
      return;
    }

    isLoading.value = true;
    errorMessage.value = '';

    try {
      final response = await _apiClient.post(
        '/auth/login',
        data: LoginRequest(username: username, password: password).toJson(),
      );

      final responseData = response.data;

      // Handle different response structures
      Map<String, dynamic> data;
      if (responseData is Map<String, dynamic>) {
        // Check for nested data structure
        if (responseData.containsKey('data')) {
          data = responseData['data'];
        } else {
          data = responseData;
        }
      } else {
        throw Exception('Invalid response format');
      }

      final loginResponse = LoginResponse.fromJson(data);

      // Save token
      await ApiClient.saveToken(loginResponse.token);

      // Update current user
      currentUser.value = loginResponse;

      // Navigate to chat list
      Get.offAll(() => const ChatListScreen());
    } on DioException catch (e) {
      _handleDioError(e);
    } catch (e) {
      errorMessage.value = 'Login failed: ${e.toString()}';
    } finally {
      isLoading.value = false;
    }
  }

  /// Handle Dio errors
  void _handleDioError(DioException e) {
    if (e.response != null) {
      final statusCode = e.response!.statusCode;
      final data = e.response!.data;

      if (data is Map<String, dynamic> && data.containsKey('message')) {
        errorMessage.value = data['message'];
      } else {
        switch (statusCode) {
          case 401:
            errorMessage.value = 'Invalid username or password';
            break;
          case 400:
            errorMessage.value = 'Invalid request';
            break;
          case 500:
            errorMessage.value = 'Server error. Please try again later.';
            break;
          default:
            errorMessage.value = 'Login failed. Please try again.';
        }
      }
    } else if (e.type == DioExceptionType.connectionTimeout) {
      errorMessage.value = 'Connection timeout. Please check your network.';
    } else if (e.type == DioExceptionType.connectionError) {
      errorMessage.value = 'Cannot connect to server. Please check your network.';
    } else {
      errorMessage.value = 'An error occurred. Please try again.';
    }
  }

  /// Logout
  Future<void> logout() async {
    isLoading.value = true;
    try {
      // Try to notify server (ignore errors)
      try {
        await _apiClient.post('/auth/logout');
      } catch (_) {
        // Ignore logout API errors
      }
    } finally {
      await ApiClient.clearToken();
      currentUser.value = null;
      isLoading.value = false;
      Get.offAllNamed('/login');
    }
  }

  /// Get current token
  Future<String?> getToken() async {
    return await ApiClient.getToken();
  }
}
