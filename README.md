# UBotHub

<p align="center">
  <strong>开放平台 — 连接聊天机器人与虚拟形象</strong>
</p>

<p align="center">
  <a href="./README_EN.md">English</a> | 简体中文
</p>

<p align="center">
  <a href="./CONTRIBUTING.md">Contributing</a> |
  <a href="./LICENSE">GPL-3.0 License</a> |
  <a href="./SECURITY.md">Security</a>
</p>

---

## 概述

UBotHub 是一个开放平台，允许用户接入自定义聊天机器人（Agent Bot），上传 3D/Live2D 虚拟人物模型并导入多种动作，实现机器人与虚拟形象的实时联动交互。

支持 AstrBot 等主流机器人框架的快速集成，提供资产管理和实时交互展示能力。

## 核心特性

- **多框架机器人接入** — 适配器模式支持 AstrBot 等 bot 框架 及通用 Webhook
- **3D 模型渲染** — 支持 VRM、glTF/GLB、FBX 格式，基于 Three.js + React Three Fiber
- **Live2D 渲染** — 支持 Cubism 2/3/4 模型，基于 PixiJS + pixi-live2d-display
- **动作映射引擎** — 关键词/情绪/正则策略链，将消息自动映射为虚拟形象动作
- **口型同步** — 文本驱动和音频驱动两种方案
- **资产管理** — 模型/动作/纹理文件的上传、处理、预览和版本管理
- **实时交互** — WebSocket 实时推送，消息到达即触发虚拟形象响应
- **架构设计** — JWT 认证、RBAC 权限、Redis 缓存、异步任务队列

## 技术栈

| 层级 | 技术 |
|------|------|
| 后端 | Go, Gin, GORM, PostgreSQL, Redis, WebSocket |
| 前端 | React 18, TypeScript, Vite, Ant Design |
| 3D 引擎 | Three.js, @react-three/fiber, @pixiv/three-vrm |
| Live2D | PixiJS, pixi-live2d-display |
| 存储 | MinIO (本地) / 阿里云 OSS (生产) |
| 队列 | asynq (Redis Streams) |
| 部署 | Docker, Docker Compose |

## 快速开始

### 环境要求

- Go 1.21+
- Node.js 18+
- pnpm 8+
- Docker & Docker Compose

### 启动开发环境

```bash
# 1. 克隆仓库
git clone git@github.com:NickCharlie/ubothub.git
cd ubothub

# 2. 启动基础设施（PostgreSQL, Redis, MinIO）
make docker-up

# 3. 启动后端（热重载）
make dev-backend

# 4. 启动前端（另一个终端）
make dev-frontend
```

### 访问地址

| 服务 | 地址 |
|------|------|
| 前端 | http://localhost:3000 |
| 后端 API | http://localhost:8080/api/v1 |
| 健康检查 | http://localhost:8080/api/v1/health |
| MinIO 控制台 | http://localhost:9001 |
| asynqmon | http://localhost:8081 |

## 项目结构

```
ubothub/
├── backend/           # Go 后端
│   ├── cmd/server/    # 应用入口
│   ├── internal/      # 核心业务代码
│   │   ├── config/    # 配置管理
│   │   ├── middleware/ # HTTP 中间件
│   │   ├── model/     # 数据模型
│   │   ├── repository/ # 数据访问层
│   │   ├── service/   # 业务逻辑层
│   │   ├── handler/   # HTTP 处理器
│   │   ├── cache/     # Redis 缓存
│   │   ├── storage/   # 对象存储
│   │   └── router/    # 路由注册
│   └── pkg/           # 公共工具包
├── frontend/          # React 前端
│   ├── src/
│   │   ├── api/       # HTTP 客户端
│   │   ├── pages/     # 页面组件
│   │   ├── stores/    # 状态管理
│   │   └── engine/    # 3D/Live2D 渲染引擎
│   └── public/
└── docker-compose.yml
```

## API 文档

启动后端后访问 Swagger UI:

```
http://localhost:8080/swagger/index.html
```

## 参与贡献

请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解如何参与项目开发。

## 许可证

本项目基于 [GPL-3.0](./LICENSE) 许可证开源。

## 安全

如发现安全漏洞，请参阅 [SECURITY.md](./SECURITY.md) 进行负责任的披露。
