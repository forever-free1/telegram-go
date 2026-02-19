import 'package:get/get.dart';

import 'core/database/database.dart';
import 'core/network/api_client.dart';
import 'core/sync/sync_controller.dart';
import 'core/websocket/websocket_service.dart';
import 'features/auth/controllers/auth_controller.dart';
import 'features/contacts/controllers/contacts_controller.dart';

/// Initialize all services
Future<void> initServices() async {
  // 1. Initialize API Client (should be first as others depend on it)
  await Get.putAsync<ApiClient>(() async {
    final api = ApiClient();
    await api.init();
    return api;
  });

  // 2. Initialize Database
  Get.put<AppDatabase>(AppDatabase(), permanent: true);

  // 3. Initialize WebSocket Service
  Get.put<WebSocketService>(WebSocketService(), permanent: true);

  // 4. Initialize Sync Controller
  Get.put<SyncController>(SyncController(), permanent: true);

  // 5. Initialize Auth Controller
  Get.put<AuthController>(AuthController(), permanent: true);

  // 6. Initialize Contacts Controller
  Get.put<ContactsController>(ContactsController(), permanent: true);
}
