# Telegram Go - 全栈即时通讯应用

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.24+-00ADD8?style=for-the-badge&logo=go" alt="Go 版本">
  <img src="https://img.shields.io/badge/Flutter-3.41+-02569B?style=for-the-badge&logo=flutter" alt="Flutter 版本">
  <img src="https://img.shields.io/badge/MySQL-8.0-4479A1?style=for-the-badge&logo=mysql" alt="MySQL 版本">
  <img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge">
</p>

> 🚀 一个现代化的即时通讯应用，基于 **Go (Gin)** 后端和 **Flutter** 前端构建，支持 Material Design 3、实时 WebSocket 通讯和 Swagger API 文档。

![项目横幅](https://via.placeholder.com/800x200/1565C0/FFFFFF?text=Telegram+Go)

## 📚 目录

- [📝 项目概述](#-项目概述)
- [🛠️ 技术栈](#️-技术栈)
- [📁 项目结构](#-项目结构)
- [🚀 快速开始](#-快速开始)
  - [后端配置](#后端配置)
  - [前端配置](#前端配置)
- [🔌 API 文档](#-api-文档)
- [🎨 界面设计](#-界面设计)
- [📖 架构指南](#-架构指南)
  - [后端架构](#后端架构)
  - [前端架构](#前端架构)
- [🔄 WebSocket 事件](#-websocket-事件)
- [🤝 贡献指南](#-贡献指南)
- [📄 许可证](#-许可证)

---

## 📝 项目概述

Telegram Go 是一个受 Telegram 和 Kelivo 优雅 UI 设计启发的全栈即时通讯应用。它提供实时消息功能，具有现代、响应迅速的界面。

### 主要功能

- ✅ **用户认证** - 使用 JWT 令牌注册/登录
- ✅ **实时通讯** - 基于 WebSocket 的即时消息
- ✅ **聊天管理** - 创建群组、添加/移除成员
- ✅ **联系人同步** - 从设备同步联系人
- ✅ **文件上传** - 分享图片和文件
- ✅ **推送通知** - 设备令牌注册用于通知
- ✅ **Material Design 3** - 现代、动态主题
- ✅ **响应式布局** - 适配手机、平板和桌面端

---

## 🛠️ 技术栈

### 后端

| 技术 | 用途 | 版本 |
|------------|---------|---------|
| **Go** | 编程语言 | 1.24+ |
| **Gin** | Web 框架 | v1.11.0 |
| **GORM** | MySQL ORM | v1.25.5 |
| **JWT** | 身份认证 | v5.2.0 |
| **WebSocket** | 实时通讯 | gorilla/websocket v1.5.1 |
| **Swagger** | API 文档 | swag v1.16.6 |
| **Snowflake** | 唯一ID生成 | v0.3.0 |
| **Zap** | 结构化日志 | v1.26.0 |

### 前端

| 技术 | 用途 | 版本 |
|------------|---------|---------|
| **Flutter** | UI 框架 | 3.41+ |
| **GetX** | 状态管理 | v4.6.6 |
| **Dio** | HTTP 客户端 | v5.7.0 |
| **Dynamic Color** | Material You 主题 | v1.8.1 |
| **Google Fonts** | 字体排版 | v6.3.3 |
| **Flutter Animate** | 动画效果 | v4.5.2 |
| **Window Manager** | 桌面窗口控制 | v0.4.3 |
| **Drift** | SQLite 数据库 | v2.22.1 |

---

## 📁 项目结构

```
telegram-go/
├── backend/                    # Go 后端
│   ├── cmd/api/               # 应用入口
│   │   └── main.go
│   ├── internal/              # 私有应用代码
│   │   ├── config/           # 配置管理
│   │   ├── database/         # 数据库连接和迁移
│   │   ├── dto/              # 数据传输对象
│   │   ├── handler/          # HTTP 请求处理
│   │   ├── middleware/       # 中间件（认证、CORS等）
│   │   ├── model/           # 数据库模型
│   │   ├── repository/       # 数据访问层
│   │   ├── service/         # 业务逻辑
│   │   └── websocket/       # WebSocket 中心和事件
│   ├── pkg/                  # 公共库
│   │   ├── crypto/          # 加密工具
│   │   └── snowflake/       # ID 生成器
│   ├── docs/                 # Swagger 文档
│   ├── config.yaml          # 应用配置
│   ├── go.mod               # Go 模块定义
│   └── go.sum               # Go 依赖
│
├── frontend/                 # Flutter 前端
│   ├── lib/
│   │   ├── core/            # 核心工具
│   │   │   ├── database/    # Drift 数据库
│   │   │   ├── layout/     # 响应式布局
│   │   │   ├── network/    # API 客户端
│   │   │   ├── sync/       # 消息同步
│   │   │   ├── theme/     # 主题配置
│   │   │   └── websocket/ # WebSocket 服务
│   │   ├── features/       # 功能模块
│   │   │   ├── auth/      # 认证功能
│   │   │   ├── chat/      # 聊天功能
│   │   │   └── contacts/  # 联系人功能
│   │   └── main.dart      # 应用入口
│   ├── web/                 # Web 平台文件
│   ├── windows/             # Windows 桌面文件
│   ├── pubspec.yaml        # Flutter 依赖
│   └── analysis_options.yaml
│
└── README.md               # 本文件
```

---

## 🚀 快速开始

### 前置要求

- **后端**: Go 1.24+, MySQL 8.0+
- **前端**: Flutter 3.41+, Dart 3.11+
- **可选**: VS Code + Flutter 插件

---

### 后端配置

#### 1. 克隆并进入目录

```bash
cd telegram-go/backend
```

#### 2. 配置数据库

编辑 `config.yaml`:

```yaml
server:
  port: "8080"          # 服务器端口
  mode: "debug"         # debug/release

database:
  host: "localhost"     # MySQL 主机
  port: "3306"         # MySQL 端口
  user: "root"         # MySQL 用户名
  password: "password" # MySQL 密码
  name: "telegram_go"  # 数据库名
  charset: "utf8mb4"

jwt:
  secret: "your-secret-key-change-in-production"
  expire_hours: 72

upload:
  path: "./uploads"
  max_size: 10485760   # 10MB
```

#### 3. 安装依赖

```bash
go mod tidy
```

#### 4. 运行服务器

```bash
go run ./cmd/api
```

服务器将会：
- 连接 MySQL 并自动迁移表
- 在端口 8080 启动 HTTP 服务器
- 在 `http://localhost:8080/swagger/index.html` 提供 Swagger 文档

#### 5. 测试 API

在浏览器中访问 `http://localhost:8080` - 会重定向到 Swagger UI。

---

### 前端配置

#### 1. 进入前端目录

```bash
cd telegram-go/frontend
```

#### 2. 安装依赖

```bash
flutter pub get
```

#### 3. 生成 Drift 数据库代码

```bash
flutter pub run build_runner build --delete-conflicting-outputs
```

#### 4. 运行应用

```bash
# Web 浏览器
flutter run -d chrome

# Windows 桌面
flutter run -d windows

# iOS 模拟器
flutter run -d iphone

# Android 模拟器
flutter run -d android
```

#### 5. 构建发布版本

```bash
# Web
flutter build web --release

# Windows
flutter build windows --release

# Android APK
flutter build apk --release
```

---

## 🔌 API 文档

### 认证接口

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/auth/register` | 注册新用户 |
| POST | `/api/auth/login` | 登录并获取 JWT |
| POST | `/api/auth/logout` | 登出当前用户 |
| GET | `/api/user/me` | 获取当前用户信息 |

### 聊天管理

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/chats` | 创建新聊天 |
| GET | `/api/chats` | 获取用户聊天列表 |
| GET | `/api/chats/:id` | 获取聊天详情 |
| POST | `/api/chats/:id/members` | 添加成员到聊天 |
| DELETE | `/api/chats/:id/members` | 移除成员 |
| GET | `/api/chats/:id/members` | 获取聊天成员 |

### 消息通讯

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/messages` | 发送消息 |
| GET | `/api/messages` | 获取聊天消息 |
| DELETE | `/api/messages/:id` | 删除消息 |
| POST | `/api/messages/ack` | 确认消息 |
| GET | `/api/sync` | 通过 SeqID 同步消息 |

### 联系人

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/contacts/sync` | 同步联系人 |
| GET | `/api/contacts` | 获取联系人列表 |
| POST | `/api/contacts` | 添加联系人 |
| DELETE | `/api/contacts/:id` | 删除联系人 |

### 其他

| 方法 | 端点 | 描述 |
|--------|----------|-------------|
| POST | `/api/upload` | 上传文件 |
| POST | `/api/device/token` | 注册推送令牌 |
| DELETE | `/api/device/token` | 注销令牌 |

### WebSocket

使用 JWT 令牌作为查询参数连接到 `ws://localhost:8080/ws`:

```
ws://localhost:8080/ws?token=<JWT_TOKEN>
```

---

## 🎨 界面设计

### Material Design 3 实现

前端实现了 Material Design 3，具有以下关键特性：

#### 1. 动态配色方案

```dart
// 使用 dynamic_color 包支持 Material You
DynamicColorBuilder(
  builder: (lightDynamic, darkDynamic) {
    // 颜色适配系统壁纸
  },
)
```

#### 2. 主题配置

- **主色**: 深蓝色 (#1565C0)
- **应用栏**: 透明背景，居中标题
- **卡片**: elevation: 0，使用 surfaceContainerLow
- **输入框**: 圆角（16px 半径）
- **导航**:
  - 移动端: NavigationBar（底部）
  - 桌面端: NavigationRail（左侧边栏）

#### 3. 聊天列表 (ChatListScreen)

- **SliverAppBar.large**: 可折叠的大标题
- **ChatTile**: 自定义 ListTile，带头像、标题、副标题
- **动画**: 使用 flutter_animate 实现交错淡入 + 滑入效果

#### 4. 聊天页面 (ChatPageScreen)

- **消息气泡**:
  - 带尾巴形状（20px 半径，4px 角落）
  - 自己消息: 主色（右对齐）
  - 他人消息: 次要容器（左对齐）
  - 最大宽度: 屏幕的 75%
- **输入栏**: 浮动胶囊样式输入框 + 发送按钮
- **Markdown 支持**: 使用 flutter_markdown 渲染

---

## 📖 架构指南

### 后端架构

后端遵循**清晰架构**，关注点分离明确：

```
Handler → Service → Repository → Model
   ↓         ↓          ↓
  DTO      业务        数据库
          逻辑        操作
```

#### 各层职责

1. **处理层** (`internal/handler/`)
   - 接收 HTTP 请求
   - 使用 DTO 验证输入
   - 调用服务方法
   - 返回 HTTP 响应

2. **服务层** (`internal/service/`)
   - 包含业务逻辑
   - 实现核心功能
   - 协调仓库
   - 处理错误

3. **仓库层** (`internal/repository/`)
   - 数据访问操作
   - 数据库查询
   - CRUD 操作

4. **模型层** (`internal/model/`)
   - 数据库表定义
   - GORM 模型
   - 关联关系

#### 关键组件

**JWT 认证中间件** (`internal/middleware/auth.go`)
- 验证 JWT 令牌
- 提取用户 ID
- 保护路由

**WebSocket 中心** (`internal/websocket/hub.go`)
- 管理连接
- 广播消息
- 处理事件

**Snowflake ID** (`pkg/snowflake/`)
- 生成唯一 ID
- Twitter Snowflake 算法
- 分布式 ID 生成

---

### 前端架构

前端使用**功能优先**组织方式，结合 GetX 进行状态管理：

```
lib/
├── core/                 # 共享工具
│   ├── database/        # Drift 数据库
│   ├── layout/          # MainLayout（响应式导航）
│   ├── network/         # API 客户端
│   ├── sync/           # 消息同步控制器
│   ├── theme/          # ThemeService
│   └── websocket/      # WebSocket 服务
├── features/            # 功能模块
│   ├── auth/           # 认证功能
│   │   ├── controllers/  # 控制器
│   │   └── screens/    # 页面
│   ├── chat/           # 聊天功能
│   │   ├── screens/    # 页面
│   │   └── widgets/   # 组件
│   └── contacts/       # 联系人功能
│       ├── controllers/  # 控制器
│       └── screens/    # 页面
└── main.dart           # 入口点
```

#### GetX 状态管理

```dart
// 示例: 创建控制器
class AuthController extends GetxController {
  final RxBool isLoading = false.obs;
  final RxString errorMessage = ''.obs;
}
```

#### Drift 数据库

项目使用 Drift（SQLite）进行本地数据存储：

```dart
// 跨平台数据库连接
import 'connection/connection.dart' // 自动选择正确的实现

// Web: 使用 WasmDatabase
// 移动端/桌面: 使用 NativeDatabase
```

#### 响应式布局

应用根据屏幕宽度自动适配：

```dart
final screenWidth = MediaQuery.of(context).size.width;
final isDesktop = screenWidth >= 600;

// 移动端: NavigationBar
// 桌面端: NavigationRail + Logo
```

---

## 🔄 WebSocket 事件

### 客户端 → 服务器

| 事件 | 数据 | 描述 |
|-------|---------|-------------|
| `auth` | `{token: string}` | 认证连接 |
| `message` | `{chatId, content}` | 发送消息 |
| `typing` | `{chatId, isTyping}` | 正在输入指示 |

### 服务器 → 客户端

| 事件 | 数据 | 描述 |
|-------|---------|-------------|
| `new_message` | `{message object}` | 接收新消息 |
| `message_ack` | `{messageId, status}` | 消息状态更新 |
| `user_online` | `{userId}` | 用户上线 |
| `user_offline` | `{userId}` | 用户离线 |

### WebSocket 连接示例 (JavaScript)

```javascript
const ws = new WebSocket('ws://localhost:8080/ws?token=YOUR_JWT');

// 认证
ws.onopen = () => {
  ws.send(JSON.stringify({
    event: 'auth',
    data: { token: 'YOUR_JWT' }
  }));
};

// 监听消息
ws.onmessage = (event) => {
  const { event: eventName, data } = JSON.parse(event.data);

  if (eventName === 'new_message') {
    console.log('新消息:', data);
  }
};
```

---

## 🤝 贡献指南

1. Fork 仓库
2. 创建功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

### 代码规范

- **后端**: 遵循 Go 标准约定，提交前运行 `go fmt`
- **前端**: 运行 `flutter analyze` 检查问题

---

## 📄 许可证

本项目基于 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件。

---

## 🙏 致谢

- [Kelivo](https://github.com/Chevey339/kelivo) - UI 设计灵感
- [Gin](https://github.com/gin-gonic/gin) - 优秀的 Web 框架
- [Flutter](https://flutter.dev) - 精美的 UI 工具包
- [Swagger](https://swagger.io) - API 文档

---

<p align="center">
  由 ❤️ 制作，Telegram Go 团队
</p>
