# Telegram Go - Full-Stack Messaging Application

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go Version">
  <img src="="Flutter-3.41+-02569B?style=for-the-badge&logo=flutter" alt="Flutter Version">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql" alt="MySQL Version">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge">
</p>

> 🚀 A modern messaging application built with **Go (Gin)** backend and **Flutter** frontend, featuring Material Design 3, real-time WebSocket communication, and Swagger API documentation.

![Project Banner](https://via.placeholder.com/800x200/1565C0/FFFFFF?text=Telegram+Go)

## 📚 Table of Contents

- [📝 Overview](#-overview)
- [🛠️ Tech Stack](#️-tech-stack)
- [📁 Project Structure](#-project-structure)
- [🚀 Getting Started](#-getting-started)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
- [🔌 API Documentation](#-api-documentation)
- [🎨 UI/UX Features](#-uiux-features)
- [📖 Architecture Guide](#-architecture-guide)
  - [Backend Architecture](#backend-architecture)
  - [Frontend Architecture](#frontend-architecture)
- [🔄 WebSocket Events](#-websocket-events)
- [🤝 Contributing](#-contributing)
- [📄 License](#-license)

---

## 📝 Overview

Telegram Go is a full-stack messaging application inspired by Telegram and Kelivo's elegant UI design. It provides real-time messaging capabilities with a modern, responsive interface.

### Key Features

- ✅ **User Authentication** - Register/Login with JWT tokens
- ✅ **Real-time Messaging** - WebSocket-based instant messaging
- ✅ **Chat Management** - Create groups, add/remove members
- ✅ **Contact Sync** - Synchronize contacts from device
- ✅ **File Upload** - Share images and files
- ✅ **Push Notifications** - Device token registration for notifications
- ✅ **Material Design 3** - Modern, dynamic theming
- ✅ **Responsive Layout** - Adapts to mobile, tablet, and desktop

---

## 🛠️ Tech Stack

### Backend

| Technology | Purpose | Version |
|------------|---------|---------|
| **Go** | Programming Language | 1.24+ |
| **Gin** | Web Framework | v1.11.0 |
| **GORM** | ORM for MySQL | v1.25.5 |
| **JWT** | Authentication | v5.2.0 |
| **WebSocket** | Real-time Communication | gorilla/websocket v1.5.1 |
| **Swagger** | API Documentation | swag v1.16.6 |
| **Snowflake** | Unique ID Generation | v0.3.0 |
| **Zap** | Structured Logging | v1.26.0 |

### Frontend

| Technology | Purpose | Version |
|------------|---------|---------|
| **Flutter** | UI Framework | 3.41+ |
| **Riverpod** | State Management | v2.6.1 |
| **Dio** | HTTP Client | v5.7.0 |
| **Dynamic Color** | Material YouTheming | v1.8.1 |
| **Google Fonts** | Typography | v6.3.3 |
| **Flutter Animate** | Animations | v4.5.2 |
| **Window Manager** | Desktop Window Control | v0.4.3 |
| **Flutter Markdown** | Markdown Rendering | v0.7.7 |

---

## 📁 Project Structure

```
telegram-go/
├── backend/                    # Go Backend
│   ├── cmd/api/               # Application entry point
│   │   └── main.go
│   ├── internal/              # Private application code
│   │   ├── config/           # Configuration management
│   │   ├── database/         # Database connection & migrations
│   │   ├── dto/              # Data Transfer Objects
│   │   ├── handler/          # HTTP request handlers
│   │   ├── middleware/       # Middleware (Auth, CORS, etc.)
│   │   ├── model/           # Database models
│   │   ├── repository/      # Data access layer
│   │   ├── service/         # Business logic
│   │   └── websocket/       # WebSocket hub & events
│   ├── pkg/                  # Public libraries
│   │   ├── crypto/          # Cryptography utilities
│   │   └── snowflake/      # ID generation
│   ├── docs/                 # Swagger documentation
│   ├── config.yaml          # Application configuration
│   ├── go.mod               # Go module definition
│   └── go.sum               # Go dependencies
│
├── frontend/                 # Flutter Frontend
│   ├── lib/
│   │   ├── core/            # Core utilities
│   │   │   ├── layout/      # Responsive layouts
│   │   │   └── theme/       # Theme configuration
│   │   ├── features/        # Feature modules
│   │   │   └── chat/        # Chat feature
│   │   │       ├── screens/ # Page screens
│   │   │       └── widgets/ # Reusable widgets
│   │   └── main.dart        # Application entry
│   ├── web/                 # Web platform files
│   ├── windows/             # Windows desktop files
│   ├── pubspec.yaml        # Flutter dependencies
│   └── analysis_options.yaml
│
└── README.md               # This file
```

---

## 🚀 Getting Started

### Prerequisites

- **Backend**: Go 1.24+, MySQL 8.0+
- **Frontend**: Flutter 3.41+, Dart 3.11+
- **Optional**: VS Code with Flutter extension

---

### Backend Setup

#### 1. Clone and Navigate

```bash
cd telegram-go/backend
```

#### 2. Configure Database

Edit `config.yaml`:

```yaml
server:
  port: "8080"          # Server port
  mode: "debug"         # debug/release

database:
  host: "localhost"     # MySQL host
  port: "3306"         # MySQL port
  user: "root"         # MySQL username
  password: "password" # MySQL password
  name: "telegram_go"  # Database name
  charset: "utf8mb4"

jwt:
  secret: "your-secret-key-change-in-production"
  expire_hours: 72

upload:
  path: "./uploads"
  max_size: 10485760   # 10MB
```

#### 3. Install Dependencies

```bash
go mod tidy
```

#### 4. Run the Server

```bash
go run ./cmd/api
```

The server will:
- Connect to MySQL and auto-migrate tables
- Start HTTP server on port 8080
- Serve Swagger docs at `http://localhost:8080/swagger/index.html`

#### 5. Test APIs

Visit `http://localhost:8080` in your browser - it redirects to Swagger UI.

---

### Frontend Setup

#### 1. Navigate to Frontend

```bash
cd telegram-go/frontend
```

#### 2. Install Dependencies

```bash
flutter pub get
```

#### 3. Run the App

```bash
# Web
flutter run -d chrome

# Windows Desktop
flutter run -d windows

# iOS Simulator
flutter run -d iphone

# Android Emulator
flutter run -d android
```

#### 4. Build for Production

```bash
# Web
flutter build web --release

# Windows
flutter build windows --release

# Android APK
flutter build apk --release
```

---

## 🔌 API Documentation

### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/register` | Register new user |
| POST | `/api/auth/login` | Login and get JWT |
| POST | `/api/auth/logout` | Logout current user |
| GET | `/api/user/me` | Get current user info |

### Chat Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/chats` | Create new chat |
| GET | `/api/chats` | Get user's chats |
| GET | `/api/chats/:id` | Get chat details |
| POST | `/api/chats Add member to chat |
| DELETE |/members` | `/api/chats/members` | Remove member |
| GET | `/api/chats/:id/members` | Get chat members |

### Messaging

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/messages` | Send message |
| GET | `/api/messages` | Get chat messages |
| DELETE | `/api/messages/:id` | Delete message |
| POST | `/api/messages/ack` | Acknowledge message |
| GET | `/api/sync` | Sync messages by SeqID |

### Contacts

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/contacts/sync` | Sync contacts |
| GET | `/api/contacts` | Get contacts list |
| POST | `/api/contacts` | Add contact |
| DELETE | `/api/contacts/:id` | Delete contact |

### Other

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/upload` | Upload file |
| POST | `/api/device/token` | Register push token |
| DELETE | `/api/device/token` | Unregister token |

### WebSocket

Connect to `ws://localhost:8080/ws` with JWT token in query parameter:

```
ws://localhost:8080/ws?token=<JWT_TOKEN>
```

---

## 🎨 UI/UX Features

### Material Design 3 Implementation

The frontend implements Material Design 3 with several key features:

#### 1. Dynamic Color Scheme

```dart
// Uses dynamic_color package for Material You support
DynamicColorBuilder(
  builder: (lightDynamic, darkDynamic) {
    // Colors adapt to system wallpaper
  },
)
```

#### 2. Theme Configuration

- **Primary Seed Color**: Deep Blue (#1565C0)
- **AppBar**: Transparent background, centered title
- **Cards**: elevation: 0, using surfaceContainerLow
- **Input Fields**: Rounded corners (16px radius)
- **Navigation**:
  - Mobile: NavigationBar (bottom)
  - Desktop: NavigationRail (left sidebar)

#### 3. Chat List (ChatListScreen)

- **SliverAppBar.large**: Collapsible large title
- **ChatTile**: Custom ListTile with avatar, title, subtitle
- **Animations**: Staggered fade-in + slide-up using flutter_animate

#### 4. Chat Page (ChatPageScreen)

- **Message Bubbles**:
  - Tail-shaped (20px radius, 4px corner)
  - Own messages: Primary color (right-aligned)
  - Others' messages: Secondary container (left-aligned)
  - Max width: 75% of screen
- **Input Bar**: Floating capsule-style input with send button
- **Markdown Support**: Renders code blocks with syntax highlighting

---

## 📖 Architecture Guide

### Backend Architecture

The backend follows **Clean Architecture** with clear separation of concerns:

```
Handler → Service → Repository → Model
   ↓         ↓          ↓
  DTO      Business    Database
          Logic       Operations
```

#### Layer Responsibilities

1. **Handler Layer** (`internal/handler/`)
   - Receives HTTP requests
   - Validates input using DTOs
   - Calls service methods
   - Returns HTTP responses

2. **Service Layer** (`internal/service/`)
   - Contains business logic
   - Implements core features
   - Coordinates repositories
   - Handles errors

3. **Repository Layer** (`internal/repository/`)
   - Data access operations
   - Database queries
   - CRUD operations

4. **Model Layer** (`internal/model/`)
   - Database table definitions
   - GORM models
   - Relationships

#### Key Components

**JWT Authentication Middleware** (`internal/middleware/auth.go`)
- Validates JWT tokens
- Extracts user ID
- Protects routes

**WebSocket Hub** (`internal/websocket/hub.go`)
- Manages connections
- Broadcasts messages
- Handles events

**Snowflake ID** (`pkg/snowflake/`)
- Generates unique IDs
- Twitter Snowflake algorithm
- Distributed ID generation

---

### Frontend Architecture

The frontend uses **Feature-First** organization with Riverpod for state management:

```
lib/
├── core/           # Shared utilities
│   ├── layout/    # MainLayout (responsive nav)
│   └── theme/    # ThemeService
├── features/      # Feature modules
│   └── chat/     # Chat feature
│       ├── screens/  # Pages
│       └── widgets/ # Components
└── main.dart     # Entry point
```

#### State Management with Riverpod

```dart
// Example: Creating a provider
final chatProvider = StateNotifierProvider<ChatNotifier, ChatState>((ref) {
  return ChatNotifier(ref.read(apiClient));
});
```

#### Responsive Layout

The app automatically adapts based on screen width:

```dart
final screenWidth = MediaQuery.of(context).size.width;
final isDesktop = screenWidth >= 600;

// Mobile: NavigationBar
// Desktop: NavigationRail with logo
```

---

## 🔄 WebSocket Events

### Client → Server

| Event | Payload | Description |
|-------|---------|-------------|
| `auth` | `{token: string}` | Authenticate connection |
| `message` | `{chatId, content}` | Send message |
| `typing` | `{chatId, isTyping}` | Typing indicator |

### Server → Client

| Event | Payload | Description |
|-------|---------|-------------|
| `new_message` | `{message object}` | Receive new message |
| `message_ack` | `{messageId, status}` | Message status update |
| `user_online` | `{userId}` | User came online |
| `user_offline` | `{userId}` | User went offline |

### Example WebSocket Connection (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=YOUR_JWT');

// Authenticate
ws.onopen = () => {
  ws.send(JSON.stringify({
    event: 'auth',
    data: { token: 'YOUR_JWT' }
  }));
};

// Listen for messages
ws.onmessage = (event) => {
  const { event: eventName, data } = JSON.parse(event.data);

  if (eventName === 'new_message') {
    console.log('New message:', data);
  }
};
```

---

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Coding Standards

- **Backend**: Follow Go standard conventions, run `go fmt` before committing
- **Frontend**: Run `flutter analyze` to check for issues

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

- [Kelivo](https://github.com/Chevey339/kelivo) - UI design inspiration
- [Gin](https://github.com/gin-gonic/gin) - Awesome web framework
- [Flutter](https://flutter.dev) - Beautiful UI toolkit
- [Swagger](https://swagger.io) - API documentation

---

<p align="center">
  Made with ❤️ by Telegram Go Team
</p>
