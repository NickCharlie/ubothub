# UBotHub

<p align="center">
  <strong>Open Platform — Connecting Chat Bots with Virtual Avatars</strong>
</p>

<p align="center">
  English | <a href="./README.md">简体中文</a>
</p>

<p align="center">
  <a href="./CONTRIBUTING.md">Contributing</a> |
  <a href="./LICENSE">GPL-3.0 License</a> |
  <a href="./SECURITY.md">Security</a>
</p>

---

## Overview

UBotHub is an open platform that enables users to connect custom chat bots (Agent Bots), upload 3D/Live2D virtual character models with motion files, and create real-time interactive experiences between bots and virtual avatars.

It supports rapid integration with mainstream bot frameworks including AstrBot, NoneBot, Wechaty, and Koishi, providing comprehensive asset management and real-time interaction capabilities.

## Key Features

- **Multi-Framework Bot Integration** — Adapter pattern supporting AstrBot, NoneBot, Wechaty, Koishi, and generic Webhook
- **3D Model Rendering** — VRM, glTF/GLB, FBX format support via Three.js + React Three Fiber
- **Live2D Rendering** — Cubism 2/3/4 model support via PixiJS + pixi-live2d-display
- **Action Mapping Engine** — Keyword/emotion/regex strategy chain that automatically maps messages to avatar actions
- **Lip Sync** — Both text-driven and audio-driven approaches
- **Asset Management** — Upload, process, preview, and manage model/motion/texture files
- **Real-time Interaction** — WebSocket push for instant avatar response on message arrival
- **Enterprise-Grade Architecture** — JWT auth, RBAC permissions, Redis caching, async task queue

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Gin, GORM, PostgreSQL, Redis, WebSocket |
| Frontend | React 18, TypeScript, Vite, Ant Design |
| 3D Engine | Three.js, @react-three/fiber, @pixiv/three-vrm |
| Live2D | PixiJS, pixi-live2d-display |
| Storage | MinIO (local) / Alibaba Cloud OSS (production) |
| Queue | asynq (Redis Streams) |
| Deployment | Docker, Docker Compose |

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- pnpm 8+
- Docker & Docker Compose

### Start Development Environment

\`\`\`bash
# 1. Clone the repository
git clone git@github.com:NickCharlie/ubothub.git
cd ubothub

# 2. Start infrastructure (PostgreSQL, Redis, MinIO)
make docker-up

# 3. Start backend (hot reload)
make dev-backend

# 4. Start frontend (in another terminal)
make dev-frontend
\`\`\`

### Access Points

| Service | URL |
|---------|-----|
| Frontend | http://localhost:3000 |
| Backend API | http://localhost:8080/api/v1 |
| Health Check | http://localhost:8080/api/v1/health |
| MinIO Console | http://localhost:9001 |
| asynqmon | http://localhost:8081 |

## Project Structure

\`\`\`
ubothub/
├── backend/           # Go backend
│   ├── cmd/server/    # Application entry
│   ├── internal/      # Core business logic
│   └── pkg/           # Shared utilities
├── frontend/          # React frontend
│   └── src/
└── docker-compose.yml
\`\`\`

## Contributing

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to contribute.

## License

This project is licensed under [GPL-3.0](./LICENSE).

## Security

If you discover a security vulnerability, please see [SECURITY.md](./SECURITY.md) for responsible disclosure.
