<h1 align="center">UBotHub</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?style=flat-square&logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/License-PolyForm_NC-blue?style=flat-square" alt="License" />
</p>

<p align="center">
  <strong>Bot + Avatar 开放平台 — 让每个 AI 机器人都有自己的虚拟形象</strong>
</p>

<p align="center">
  <a href="./README_EN.md">English</a> | 简体中文
</p>

<p align="center">
  <a href="./CONTRIBUTING.md">参与贡献</a> ·
  <a href="./SECURITY.md">安全策略</a> ·
  <a href="./LICENSE">开源许可</a>
</p>

---

## 我们是什么

**UBotHub** 是一个面向 AI 机器人生态的开放平台。我们提供从机器人接入、虚拟形象管理到实时交互的全链路基础设施，让开发者可以在几分钟内将自己的聊天机器人与 3D/Live2D 虚拟角色绑定，创造有温度的 AI 交互体验。

无论你是个人开发者、企业客户还是内容创作者，UBotHub 都能让你的 AI 从「纯文字对话」进化为「有形象、有表情、有动作」的虚拟伙伴。

## 核心能力

### 对用户

| 能力 | 说明 |
|------|------|
| **一键接入 Bot** | 注册平台账号，创建 Bot，获取 Access Token，即可通过 多种渠道 接入你的聊天机器人 |
| **上传虚拟形象** | 支持 VRM、glTF/GLB、FBX（3D）和 Cubism 2/3/4（Live2D）等主流格式 |
| **实时互动** | 用户发送消息 → Bot 回复 → 虚拟形象自动做出对应表情和动作，全程 WebSocket 实时推送 |
| **Bot 广场** | 浏览和体验其他开发者创建的 Bot，发现有趣的 AI 角色 |

### 对 Bot 开发者

| 能力 | 说明 |
|------|------|
| **多框架适配** | 原生支持 AstrBot 等主流框架，也支持通用 Webhook |
| **收益变现** | 设置按次计费或订阅制定价，平台 80/20 分成直接结算到你的账户 |
| **OpenAPI 文档** | 完整的 Swagger API 文档，集成只需几行代码 |
| **实时监控** | WebSocket 连接状态、消息吞吐量、调用计费一目了然 |

### 对合作伙伴

| 优势 | 说明 |
|------|------|
| **差异化定位** | 业内首个「Bot + 虚拟形象」开放平台，解决 AI 交互「只有文字没有形象」的痛点 |
| **开发者生态** | 通过 Bot 广场、收益分成、API 开放能力吸引 Bot 开发者入驻，形成内容飞轮 |
| **技术壁垒** | 自研动作映射引擎、高并发 WebSocket 网关、服务商模式支付体系 |
| **变现能力** | 平台抽成 + 虚拟形象资产交易 + 企业级 SaaS 定制，多元化收入模型 |
| **生产级架构** | 微服务分离部署、断路器容错、异步任务队列、水平扩展能力 |

## 技术亮点

```
┌──────────────────────────────────────────────────────────────────┐
│                         UBotHub 架构                             │
├──────────────┬───────────────────┬───────────────────────────────┤
│  Frontend    │  React 19 + Vite  │ Glassmorphic Dark UI          │
│              │  Tailwind CSS     │ Framer Motion 动画            │
│              │  Three.js         │ 3D/Live2D 实时渲染            │
├──────────────┼───────────────────┼───────────────────────────────┤
│  API Layer   │  Gin Framework    │ JWT Auth + RBAC               │
│              │  Swagger/OpenAPI  │ 52+ 个 RESTful 端点            │
│              │  WebSocket Hub    │ 万级并发连接                    │
├──────────────┼───────────────────┼───────────────────────────────┤
│  Business    │  Bot 管理/网关     │ 适配器工厂模式                  │
│              │  钱包/计费/分成    │ 服务商模式支付（微信/支付宝）     │
│              │  资产/形象管理     │ 动作映射引擎                    │
├──────────────┼───────────────────┼───────────────────────────────┤
│  Infra       │  PostgreSQL 16    │ GORM + 软删除                  │
│              │  Redis 7          │ 缓存 / 限流 / 队列              │
│              │  MinIO / OSS      │ S3 兼容对象存储                  │
│              │  asynq            │ 异步任务（邮件/资产处理）         │
│              │  Docker Compose   │ 生产 / 调试双环境部署            │
└──────────────┴───────────────────┴───────────────────────────────┘
```

## 快速开始

### 环境要求

- Go 1.25+
- Node.js 22+ (LTS)
- pnpm 9+
- Docker & Docker Compose

### 方式一：本地开发（推荐）

```bash
# 克隆仓库
git clone git@github.com:NickCharlie/ubothub.git
cd ubothub

# 启动基础设施（PostgreSQL, Redis, MinIO, asynqmon）
make docker-up

# 终端 1：启动后端（air 热重载）
make dev-backend

# 终端 2：启动前端（Vite HMR）
make dev-frontend
```

### 方式二：Docker 调试环境（全栈容器化）

```bash
# 一键启动全栈（含热重载 + Delve 远程调试）
make docker-debug-up

# 查看日志
make docker-debug-logs
```

### 方式三：Docker 生产环境

```bash
# 配置环境变量
cp .env.example .env
vim .env

# 构建并启动（多阶段构建 + nginx 前端 + API/Worker 分离）
make docker-prod-up
```

### 服务地址

| 服务 | 本地开发 | Docker 调试 |
|------|---------|------------|
| 前端 | http://localhost:5173 | http://localhost:5173 |
| 后端 API | http://localhost:8080/api/v1 | http://localhost:8080/api/v1 |
| Swagger 文档 | http://localhost:8080/swagger/index.html | 同左 |
| MinIO 控制台 | http://localhost:9001 | http://localhost:9001 |
| asynqmon | http://localhost:8081 | http://localhost:8081 |
| Delve 调试 | — | localhost:2345 |

## 项目结构

```
ubothub/
├── backend/                    # Go 后端服务
│   ├── cmd/server/             # 应用入口（API / Worker / All 模式）
│   ├── internal/
│   │   ├── adapter/            # Bot 框架适配器（AstrBot, Webhook）
│   │   ├── config/             # 配置管理（Viper + 环境变量）
│   │   ├── handler/            # HTTP 处理器（12 个模块）
│   │   ├── middleware/         # 中间件（JWT, CORS, CSRF, 限流, 安全头）
│   │   ├── model/              # GORM 数据模型
│   │   ├── payment/            # 支付集成（微信/支付宝服务商模式）
│   │   ├── repository/         # 数据访问层
│   │   ├── router/             # 路由注册
│   │   ├── service/            # 业务逻辑层
│   │   ├── ws/                 # WebSocket Hub（gorilla/websocket）
│   │   └── queue/              # 异步任务处理器
│   ├── pkg/                    # 公共工具（JWT, 日志, 邮件, HTTP 连接池）
│   ├── configs/                # 配置文件（开发/调试/生产）
│   └── docs/                   # Swagger 自动生成文档
├── frontend/                   # React 前端
│   └── src/
│       ├── api/                # API 客户端
│       ├── components/         # 通用组件（Layout, Sidebar）
│       ├── pages/              # 页面（Auth, Dashboard, Bot, Avatar, Asset）
│       ├── stores/             # Zustand 状态管理
│       └── styles/             # 全局样式 + Glassmorphic 设计系统
├── .github/                    # CI/CD（Commit Lint + AI Issue Triage）
├── docker-compose.yml          # 本地基础设施
├── docker-compose.debug.yml    # Docker 调试环境
├── docker-compose.prod.yml     # Docker 生产环境
└── Makefile                    # 开发/构建/部署命令
```

## API 文档

启动后端后访问 **Swagger UI**：

```
http://localhost:8080/swagger/index.html
```

覆盖 52+ 个 API 端点，包括：认证、用户、Bot 管理、资产管理、虚拟形象、钱包计费、支付、网关、WebSocket、管理后台。

## 路线图

- [x] 用户认证系统（JWT + 邮箱验证 + 密码重置）
- [x] Bot 管理与网关（CRUD + Webhook + 消息推送）
- [x] 3D/Live2D 资产与虚拟形象管理
- [x] 钱包、计费、收益分成系统
- [x] 微信/支付宝服务商模式支付
- [x] WebSocket 实时消息网关
- [x] 管理后台 API
- [x] Swagger/OpenAPI 文档
- [x] Docker 多环境部署（开发/调试/生产）
- [ ] 前端 Bot 广场与聊天界面
- [ ] 3D/Live2D 实时渲染引擎
- [ ] 官方支付分账 API 集成
- [ ] 创作者提现系统
- [ ] 管理后台前端

## 参与贡献

欢迎提交 Pull Request！请阅读 [CONTRIBUTING.md](./CONTRIBUTING.md) 了解开发规范和提交约定。

## 许可证

本项目基于 [PolyForm Noncommercial 1.0.0](./LICENSE) 许可证开源。

## 安全

如发现安全漏洞，请参阅 [SECURITY.md](./SECURITY.md) 进行负责任的披露。
