<h1 align="center">UBotHub</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.25-00ADD8?style=flat-square&logo=go" alt="Go" />
  <img src="https://img.shields.io/badge/React-19-61DAFB?style=flat-square&logo=react" alt="React" />
  <img src="https://img.shields.io/badge/TypeScript-5.9-3178C6?style=flat-square&logo=typescript" alt="TypeScript" />
  <img src="https://img.shields.io/badge/PostgreSQL-16-4169E1?style=flat-square&logo=postgresql" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/License-PolyForm_NC-blue?style=flat-square" alt="License" />
</p>

<p align="center">
  <strong>Bot + Avatar Open Platform — Give Every AI Bot Its Own Virtual Character</strong>
</p>

<p align="center">
  English | <a href="./README.md">简体中文</a>
</p>

<p align="center">
  <a href="./CONTRIBUTING.md">Contributing</a> ·
  <a href="./SECURITY.md">Security</a> ·
  <a href="./LICENSE">License</a>
</p>

---

## What Is UBotHub

**UBotHub** is an open platform for the AI bot ecosystem. We provide end-to-end infrastructure — from bot integration and virtual avatar management to real-time interaction — enabling developers to bind their chatbots with 3D/Live2D virtual characters in minutes and create engaging AI experiences with personality.

Whether you're an indie developer, enterprise customer, or content creator, UBotHub transforms your AI from "text-only chat" into a virtual assistant with expressions, gestures, and character.

## What We Offer

### For Users

| Capability | Description |
|------------|-------------|
| **One-Click Bot Integration** | Register, create a Bot, get an Access Token, and connect your chatbot via Webhook |
| **Upload Virtual Avatars** | Support for VRM, glTF/GLB, FBX (3D) and Cubism 2/3/4 (Live2D) formats |
| **Real-Time Interaction** | User sends message → Bot replies → Avatar automatically performs matching expressions and actions, all via WebSocket |
| **Bot Plaza** | Browse and experience bots created by other developers, discover interesting AI characters |

### For Bot Developers

| Capability | Description |
|------------|-------------|
| **Multi-Framework Support** | Native adapters for AstrBot, NoneBot, Wechaty, Koishi, and generic Webhook |
| **Revenue Monetization** | Set per-call or subscription pricing; 80/20 revenue split settled directly to your account |
| **OpenAPI Documentation** | Complete Swagger API docs — integration takes just a few lines of code |
| **Real-Time Monitoring** | WebSocket connection status, message throughput, and billing metrics at a glance |

### For Investors

| Advantage | Description |
|-----------|-------------|
| **Differentiated Positioning** | The first open platform combining "Bot + Virtual Avatar", solving the pain point of AI interactions being text-only |
| **Developer Ecosystem** | Bot Plaza, revenue sharing, and open API capabilities attract bot developers, creating a content flywheel |
| **Technical Moat** | Proprietary action mapping engine, high-concurrency WebSocket gateway, service provider payment system |
| **Monetization Model** | Platform commission + virtual avatar asset marketplace + enterprise SaaS customization — diversified revenue streams |
| **Production-Grade Architecture** | Microservice-separated deployment, circuit breaker fault tolerance, async task queues, horizontal scaling |

## Technical Highlights

```
┌──────────────────────────────────────────────────────────────────┐
│                      UBotHub Architecture                        │
├──────────────┬───────────────────┬───────────────────────────────┤
│  Frontend    │  React 19 + Vite  │ Glassmorphic Dark UI          │
│              │  Tailwind CSS     │ Framer Motion Animations      │
│              │  Three.js         │ 3D / Live2D Real-Time Render  │
├──────────────┼───────────────────┼───────────────────────────────┤
│  API Layer   │  Gin Framework    │ JWT Auth + RBAC               │
│              │  Swagger/OpenAPI  │ 52+ RESTful Endpoints         │
│              │  WebSocket Hub    │ 10K+ Concurrent Connections   │
├──────────────┼───────────────────┼───────────────────────────────┤
│  Business    │  Bot Management   │ Adapter Factory Pattern       │
│              │  Wallet / Billing │ Service Provider Payments     │
│              │  Asset / Avatar   │ Action Mapping Engine         │
├──────────────┼───────────────────┼───────────────────────────────┤
│  Infra       │  PostgreSQL 16    │ GORM + Soft Delete            │
│              │  Redis 7          │ Cache / Rate Limit / Queue    │
│              │  MinIO / OSS      │ S3-Compatible Object Storage  │
│              │  asynq            │ Async Tasks (Email / Assets)  │
│              │  Docker Compose   │ Prod / Debug Dual Deployment  │
└──────────────┴───────────────────┴───────────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 22+ (LTS)
- pnpm 9+
- Docker & Docker Compose

### Option 1: Local Development (Recommended)

```bash
# Clone the repository
git clone git@github.com:NickCharlie/ubothub.git
cd ubothub

# Start infrastructure (PostgreSQL, Redis, MinIO, asynqmon)
make docker-up

# Terminal 1: Start backend (air hot-reload)
make dev-backend

# Terminal 2: Start frontend (Vite HMR)
make dev-frontend
```

### Option 2: Docker Debug Environment (Full-Stack Containerized)

```bash
# Launch full stack with hot-reload + Delve remote debugger
make docker-debug-up

# View logs
make docker-debug-logs
```

### Option 3: Docker Production Environment

```bash
# Configure environment variables
cp .env.example .env
vim .env

# Build and launch (multi-stage build + nginx frontend + API/Worker separation)
make docker-prod-up
```

### Service Endpoints

| Service | Local Dev | Docker Debug |
|---------|-----------|-------------|
| Frontend | http://localhost:5173 | http://localhost:5173 |
| Backend API | http://localhost:8080/api/v1 | http://localhost:8080/api/v1 |
| Swagger Docs | http://localhost:8080/swagger/index.html | Same |
| MinIO Console | http://localhost:9001 | http://localhost:9001 |
| asynqmon | http://localhost:8081 | http://localhost:8081 |
| Delve Debugger | — | localhost:2345 |

## Project Structure

```
ubothub/
├── backend/                    # Go backend service
│   ├── cmd/server/             # Entry point (API / Worker / All modes)
│   ├── internal/
│   │   ├── adapter/            # Bot framework adapters (AstrBot, Webhook)
│   │   ├── config/             # Configuration (Viper + env vars)
│   │   ├── handler/            # HTTP handlers (12 modules)
│   │   ├── middleware/         # Middleware (JWT, CORS, CSRF, rate limit, security)
│   │   ├── model/              # GORM data models
│   │   ├── payment/            # Payment integration (WeChat/Alipay service provider)
│   │   ├── repository/         # Data access layer
│   │   ├── router/             # Route registration
│   │   ├── service/            # Business logic layer
│   │   ├── ws/                 # WebSocket Hub (gorilla/websocket)
│   │   └── queue/              # Async task handlers
│   ├── pkg/                    # Shared utilities (JWT, logger, email, HTTP pool)
│   ├── configs/                # Config files (dev / debug / prod)
│   └── docs/                   # Auto-generated Swagger docs
├── frontend/                   # React frontend
│   └── src/
│       ├── api/                # API client modules
│       ├── components/         # Shared components (Layout, Sidebar)
│       ├── pages/              # Pages (Auth, Dashboard, Bot, Avatar, Asset)
│       ├── stores/             # Zustand state management
│       └── styles/             # Global styles + Glassmorphic design system
├── .github/                    # CI/CD (Commit Lint + AI Issue Triage)
├── docker-compose.yml          # Local infrastructure
├── docker-compose.debug.yml    # Docker debug environment
├── docker-compose.prod.yml     # Docker production environment
└── Makefile                    # Dev / build / deploy commands
```

## API Documentation

After starting the backend, visit **Swagger UI**:

```
http://localhost:8080/swagger/index.html
```

Covers 52+ API endpoints including: Authentication, Users, Bot Management, Asset Management, Virtual Avatars, Wallet & Billing, Payments, Gateway, WebSocket, and Admin Console.

## Roadmap

- [x] User authentication (JWT + email verification + password reset)
- [x] Bot management and gateway (CRUD + Webhook + message push)
- [x] 3D/Live2D asset and avatar management
- [x] Wallet, billing, and revenue sharing system
- [x] WeChat/Alipay service provider payment integration
- [x] WebSocket real-time messaging gateway
- [x] Admin management API
- [x] Swagger/OpenAPI documentation
- [x] Docker multi-environment deployment (dev / debug / prod)
- [ ] Frontend Bot Plaza and chat interface
- [ ] 3D/Live2D real-time rendering engine
- [ ] Official payment profit-sharing API
- [ ] Creator withdrawal system
- [ ] Admin console frontend

## Contributing

Pull requests welcome! Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for development standards and commit conventions.

## License

This project is licensed under [PolyForm Noncommercial 1.0.0](./LICENSE).

## Security

If you discover a security vulnerability, please see [SECURITY.md](./SECURITY.md) for responsible disclosure.
