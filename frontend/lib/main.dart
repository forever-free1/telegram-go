import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:get/get.dart';
import 'package:window_manager/window_manager.dart';
import 'package:dynamic_color/dynamic_color.dart';

import 'core/theme/theme_service.dart';
import 'core/layout/main_layout.dart';
import 'core/network/api_client.dart';
import 'core/database/database_service.dart';
import 'core/sync/sync_controller.dart';
import 'features/auth/controllers/auth_controller.dart';
import 'features/auth/screens/login_page.dart';
import 'features/chat/screens/chat_list_screen.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // Initialize services
  await initServices();

  // Initialize window manager for desktop
  await windowManager.ensureInitialized();

  const windowOptions = WindowOptions(
    size: Size(1200, 800),
    minimumSize: Size(400, 600),
    center: true,
    backgroundColor: Colors.transparent,
    skipTaskbar: false,
    titleBarStyle: TitleBarStyle.hidden,
    title: 'Telegram Go',
  );

  await windowManager.waitUntilReadyToShow(windowOptions, () async {
    await windowManager.show();
    await windowManager.focus();
  });

  // Set system UI overlay style
  SystemChrome.setSystemUIOverlayStyle(
    const SystemUiOverlayStyle(
      statusBarColor: Colors.transparent,
      statusBarIconBrightness: Brightness.dark,
      systemNavigationBarColor: Colors.transparent,
    ),
  );

  runApp(const MyApp());
}

/// Initialize all GetX services
Future<void> initServices() async {
  // Initialize API Client
  await Get.putAsync(() => ApiClient().init());

  // Initialize Database (Isar)
  await Get.putAsync(() => DatabaseService().init());

  // Initialize Sync Controller
  Get.put(SyncController());

  // Initialize Auth Controller
  Get.put(AuthController());
}

class MyApp extends StatelessWidget {
  const MyApp({super.key});

  @override
  Widget build(BuildContext context) {
    return DynamicColorBuilder(
      builder: (ColorScheme? lightDynamic, ColorScheme? darkDynamic) {
        final lightTheme = ThemeService.lightThemeWithColorScheme(
          lightDynamic ?? ColorScheme.fromSeed(seedColor: const Color(0xFF1565C0), brightness: Brightness.light),
        );
        final darkTheme = ThemeService.darkThemeWithColorScheme(
          darkDynamic ?? ColorScheme.fromSeed(seedColor: const Color(0xFF1565C0), brightness: Brightness.dark),
        );

        return GetMaterialApp(
          title: 'Telegram Go',
          debugShowCheckedModeBanner: false,
          theme: lightTheme,
          darkTheme: darkTheme,
          themeMode: ThemeMode.system,
          initialRoute: '/',
          getPages: [
            GetPage(
              name: '/',
              page: () => const AuthWrapper(),
            ),
            GetPage(
              name: '/login',
              page: () => const LoginPage(),
            ),
            GetPage(
              name: '/home',
              page: () => const MainLayout(),
            ),
            GetPage(
              name: '/chats',
              page: () => const ChatListScreen(),
            ),
          ],
        );
      },
    );
  }
}

/// Auth Wrapper - Checks login status and redirects
class AuthWrapper extends StatelessWidget {
  const AuthWrapper({super.key});

  @override
  Widget build(BuildContext context) {
    // Check if user is logged in
    return FutureBuilder<bool>(
      future: ApiClient.hasToken(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Scaffold(
            body: Center(
              child: CircularProgressIndicator(),
            ),
          );
        }

        if (snapshot.data == true) {
          return const MainLayout();
        }

        return const LoginPage();
      },
    );
  }
}
