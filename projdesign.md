# UBotHub 完整开发方案

## Context

构建一个开放平台,核心功能是允许用户接入自定义聊天机器人(Agent Bot),同时支持上传 3D/Live2D 人物模型并导入多种动作,最终实现机器人与虚拟形象的联动交互。平台支持 AstrBot、NoneBot、Wechaty、Koishi 等主流机器人框架的快速集成,提供完善的资产管理和实时交互展示能力。

**技术约束**: 不使用 Java/Python。后端 Go,前端 React + TypeScript,pnpm 管理。

---

## 1. 技术选型

### 后端 (Go)

| 组件 | 技术 | 说明 |
|------|------|------|
| Web 框架 | Gin | 高性能, 丰富中间件生态, 企业级广泛采用 |
| ORM | GORM | Go 最成熟的 ORM, 支持迁移/钩子/预加载 |
| 数据库 | PostgreSQL 16 | JSONB 支持, 全文搜索 |
| 缓存 | Redis 7.x (go-redis/v9) | 缓存 + Pub/Sub + 分布式锁 |
| WebSocket | gorilla/websocket | Go 标准 WebSocket 库, 高并发连接 |
| 任务队列 | asynq | Redis-based 异步任务队列(Go 原生) |
| 对象存储 | MinIO (本地) / 阿里云 OSS (生产) | Strategy 模式切换, 统一 S3 兼容接口 |
| 认证 | golang-jwt/jwt/v5 | JWT 签发与验证 |
| 配置管理 | Viper | 支持 YAML/ENV/远程配置 |
| 依赖注入 | Wire (Google) | 编译时 DI, 零运行时开销 |
| API 文档 | swaggo/swag | 从注解自动生成 Swagger/OpenAPI |
| 日志 | zap (Uber) | 高性能结构化日志 |
| 数据校验 | go-playground/validator | 结构体标签验证 |
| 监控 | prometheus/client_golang | Prometheus metrics 暴露 |
| 热重载 | air | 开发环境文件监听与自动重编译 |
| 测试 | testify + gomock | 断言库 + Mock 生成 |

### 前端 (React + TypeScript)

| 组件 | 技术 | 说明 |
|------|------|------|
| 框架 | React 18 + TypeScript 5 | |
| 构建 | Vite 5.x | 快速 HMR |
| UI 组件 | Ant Design 5.x | 企业级, 国际化完善 |
| 3D 渲染 | @react-three/fiber (R3F) | React 声明式 Three.js |
| 3D 辅助 | @react-three/drei | R3F 常用 helper 集合 |
| VRM | @pixiv/three-vrm 3.x | VRM 模型加载渲染 |
| Live2D | pixi-live2d-display + PixiJS | Cubism 2/3/4 支持 |
| 状态管理 | Zustand | 轻量, TypeScript 友好 |
| 路由 | React Router v6 | SPA 路由 |
| HTTP | Axios | API 请求 |
| WebSocket | socket.io-client | 实时通信客户端 |
| 图表 | ECharts (echarts-for-react) | 数据可视化 |
| 表单 | React Hook Form + Zod | 表单校验 |
| 代码规范 | ESLint + Prettier | Google Style |
| 测试 | Vitest + React Testing Library | 单元/集成测试 |

### 基础设施

| 组件 | 技术 | 说明 |
|------|------|------|
| 容器化 | Docker + Docker Compose | 开发/部署一致性 |
| CI/CD | GitHub Actions | 自动测试/构建/部署 |
| 监控 | Prometheus + Grafana | 系统监控与告警 |
| 包管理 | pnpm | 前端依赖管理 |
| Go 模块 | Go Modules | 后端依赖管理 |

---

## 2. 项目目录结构

```
ubothub/
├── docker-compose.yml              # 开发环境编排
├── docker-compose.prod.yml         # 生产环境编排
├── Makefile                        # 构建/运行/测试快捷命令
├── .github/
│   └── workflows/
│       ├── ci.yml                  # 持续集成(lint/test/build)
│       └── deploy.yml              # 部署流水线
│
├── backend/                        # Go 后端
│   ├── cmd/
│   │   └── server/
│   │       └── main.go             # 应用入口
│   ├── internal/                   # 私有应用代码
│   │   ├── config/                 # 配置加载
│   │   │   └── config.go
│   │   ├── middleware/             # Gin 中间件
│   │   │   ├── auth.go             # JWT 认证中间件
│   │   │   ├── cors.go             # CORS 配置
│   │   │   ├── logger.go           # 请求日志
│   │   │   ├── ratelimit.go        # 限流
│   │   │   └── recovery.go         # Panic 恢复
│   │   ├── model/                  # GORM 数据模型
│   │   │   ├── user.go
│   │   │   ├── bot.go
│   │   │   ├── asset.go
│   │   │   ├── avatar.go
│   │   │   ├── message_log.go
│   │   │   └── action_template.go
│   │   ├── repository/             # 数据访问层(Repository Pattern)
│   │   │   ├── user_repo.go
│   │   │   ├── bot_repo.go
│   │   │   ├── asset_repo.go
│   │   │   ├── avatar_repo.go
│   │   │   └── message_log_repo.go
│   │   ├── service/                # 业务逻辑层
│   │   │   ├── auth_service.go
│   │   │   ├── user_service.go
│   │   │   ├── bot_service.go
│   │   │   ├── asset_service.go
│   │   │   ├── avatar_service.go
│   │   │   ├── interaction_service.go
│   │   │   └── storage_service.go
│   │   ├── handler/                # HTTP Handler 层(Controller)
│   │   │   ├── auth_handler.go
│   │   │   ├── user_handler.go
│   │   │   ├── bot_handler.go
│   │   │   ├── asset_handler.go
│   │   │   ├── avatar_handler.go
│   │   │   └── interaction_handler.go
│   │   ├── router/                 # 路由注册
│   │   │   └── router.go
│   │   ├── adapter/                # 机器人适配器(Adapter Pattern)
│   │   │   ├── adapter.go          # 接口定义
│   │   │   ├── factory.go          # 适配器工厂
│   │   │   ├── astrbot.go          # AstrBot 适配器
│   │   │   ├── nonebot.go          # NoneBot 适配器
│   │   │   ├── wechaty.go          # Wechaty 适配器
│   │   │   ├── koishi.go           # Koishi 适配器
│   │   │   └── webhook.go          # 通用 Webhook 适配器
│   │   ├── gateway/                # WebSocket 网关
│   │   │   ├── hub.go              # 连接管理中心
│   │   │   ├── client.go           # 客户端连接抽象
│   │   │   ├── handler.go          # 消息处理器
│   │   │   └── protocol.go         # 事件协议定义
│   │   ├── interaction/            # 交互引擎
│   │   │   ├── engine.go           # 交互引擎核心
│   │   │   ├── action_mapper.go    # 动作映射器
│   │   │   ├── state_machine.go    # 动画状态机
│   │   │   └── strategy/           # 映射策略(Strategy Pattern)
│   │   │       ├── strategy.go     # 策略接口
│   │   │       ├── keyword.go      # 关键词策略
│   │   │       ├── emotion.go      # 情绪分析策略
│   │   │       ├── regex.go        # 正则策略
│   │   │       └── chain.go        # 策略链
│   │   ├── queue/                  # 异步任务
│   │   │   ├── client.go           # asynq 客户端
│   │   │   ├── server.go           # asynq 工作进程
│   │   │   └── tasks/
│   │   │       ├── asset_process.go    # 资产处理任务
│   │   │       ├── thumbnail_gen.go    # 缩略图生成
│   │   │       └── message_dispatch.go # 消息分发任务
│   │   ├── event/                  # 事件总线(Observer Pattern)
│   │   │   ├── bus.go              # 事件总线
│   │   │   └── types.go            # 事件类型定义
│   │   └── dto/                    # 数据传输对象
│   │       ├── request/
│   │       │   ├── auth_req.go
│   │       │   ├── bot_req.go
│   │       │   ├── asset_req.go
│   │       │   └── avatar_req.go
│   │       └── response/
│   │           ├── common.go       # 统一响应结构
│   │           ├── auth_resp.go
│   │           ├── bot_resp.go
│   │           ├── asset_resp.go
│   │           └── avatar_resp.go
│   ├── pkg/                        # 可复用的公共包
│   │   ├── errcode/                # 错误码定义
│   │   │   └── errcode.go
│   │   ├── hash/                   # 密码哈希工具
│   │   │   └── bcrypt.go
│   │   ├── token/                  # JWT 工具
│   │   │   └── jwt.go
│   │   └── response/               # 统一响应封装
│   │       └── response.go
│   ├── configs/
│   │   ├── config.yaml             # 默认配置
│   │   └── config.prod.yaml        # 生产配置
│   ├── migrations/                 # 数据库迁移(GORM AutoMigrate 或 golang-migrate)
│   ├── docs/                       # swag 自动生成的 API 文档
│   ├── wire.go                     # Wire DI 定义
│   ├── wire_gen.go                 # Wire 生成代码
│   ├── go.mod
│   ├── go.sum
│   └── Dockerfile
│
├── frontend/                       # React 前端
│   ├── src/
│   │   ├── main.tsx                # 应用入口
│   │   ├── App.tsx                 # 根组件
│   │   ├── api/                    # API 请求层
│   │   │   ├── http.ts             # Axios 实例(拦截器, baseURL)
│   │   │   ├── auth.api.ts
│   │   │   ├── bot.api.ts
│   │   │   ├── asset.api.ts
│   │   │   ├── avatar.api.ts
│   │   │   └── interaction.api.ts
│   │   ├── components/             # 通用组件
│   │   │   ├── layout/
│   │   │   │   ├── AppLayout.tsx
│   │   │   │   ├── AppHeader.tsx
│   │   │   │   └── AppSidebar.tsx
│   │   │   ├── common/             # 基础 UI 组件封装
│   │   │   └── feedback/           # 反馈组件
│   │   ├── hooks/                  # 自定义 Hooks
│   │   │   ├── useAuth.ts
│   │   │   ├── useWebSocket.ts
│   │   │   └── useNotification.ts
│   │   ├── router/
│   │   │   └── index.tsx
│   │   ├── stores/                 # Zustand 状态管理
│   │   │   ├── authStore.ts
│   │   │   ├── botStore.ts
│   │   │   ├── assetStore.ts
│   │   │   └── avatarStore.ts
│   │   ├── pages/                  # 页面视图
│   │   │   ├── auth/
│   │   │   │   ├── LoginPage.tsx
│   │   │   │   └── RegisterPage.tsx
│   │   │   ├── dashboard/
│   │   │   │   └── DashboardPage.tsx
│   │   │   ├── bot/
│   │   │   │   ├── BotListPage.tsx
│   │   │   │   ├── BotDetailPage.tsx
│   │   │   │   └── BotConfigPage.tsx
│   │   │   ├── asset/
│   │   │   │   ├── AssetLibraryPage.tsx
│   │   │   │   ├── ModelUploadPage.tsx
│   │   │   │   └── MotionUploadPage.tsx
│   │   │   ├── avatar/
│   │   │   │   ├── AvatarEditorPage.tsx
│   │   │   │   └── AvatarPreviewPage.tsx
│   │   │   └── interaction/
│   │   │       └── InteractionPage.tsx
│   │   ├── engine/                 # 3D/Live2D 渲染引擎
│   │   │   ├── three/              # Three.js 相关
│   │   │   │   ├── SceneContainer.tsx    # R3F Canvas 容器
│   │   │   │   ├── ModelViewer.tsx       # 通用模型查看器
│   │   │   │   ├── VRMAvatar.tsx         # VRM 角色组件
│   │   │   │   ├── GLTFAvatar.tsx        # GLTF 角色组件
│   │   │   │   ├── FBXAvatar.tsx         # FBX 角色组件
│   │   │   │   └── controls/
│   │   │   │       ├── OrbitControls.tsx  # 轨道控制
│   │   │   │       └── CameraRig.tsx     # 相机控制
│   │   │   ├── animation/          # 动画系统
│   │   │   │   ├── AnimationController.ts
│   │   │   │   ├── AnimationMixer.ts     # 动画混合器封装
│   │   │   │   ├── LipSyncController.ts  # 口型同步
│   │   │   │   ├── EmotionController.ts  # 表情控制
│   │   │   │   └── StateMachine.ts       # 动画状态机
│   │   │   ├── loaders/            # 模型加载器
│   │   │   │   ├── LoaderFactory.ts      # 加载器工厂
│   │   │   │   ├── VRMLoader.ts
│   │   │   │   ├── GLTFLoaderWrapper.ts
│   │   │   │   ├── FBXLoaderWrapper.ts
│   │   │   │   └── MotionLoader.ts       # BVH/VMD 动作加载
│   │   │   ├── live2d/             # Live2D 渲染
│   │   │   │   ├── Live2DCanvas.tsx      # Live2D 画布组件
│   │   │   │   ├── Live2DRenderer.ts     # 渲染器封装
│   │   │   │   ├── Live2DAnimator.ts     # 动画控制
│   │   │   │   └── Live2DLipSync.ts      # 口型同步
│   │   │   └── hooks/              # 渲染引擎 Hooks
│   │   │       ├── useModelLoader.ts
│   │   │       ├── useAnimation.ts
│   │   │       ├── useLipSync.ts
│   │   │       └── useLive2D.ts
│   │   ├── types/                  # TypeScript 类型定义
│   │   │   ├── bot.ts
│   │   │   ├── asset.ts
│   │   │   ├── avatar.ts
│   │   │   ├── event.ts
│   │   │   └── api.ts
│   │   ├── constants/              # 常量
│   │   │   ├── eventTypes.ts
│   │   │   ├── modelFormats.ts
│   │   │   └── actionTypes.ts
│   │   ├── utils/                  # 工具函数
│   │   └── styles/
│   ├── public/
│   ├── package.json
│   ├── pnpm-lock.yaml
│   ├── vite.config.ts
│   ├── tsconfig.json
│   └── Dockerfile
│
├── projdesign.md                   # 项目设计文档(本文件)
├── pnpm-workspace.yaml             # pnpm 工作区(仅前端)
├── .eslintrc.cjs
├── .prettierrc
└── README.md
```

---

## 3. 数据库设计

### 3.1 用户表 (users)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) CUID | 主键 |
| email | VARCHAR(255) UNIQUE | 邮箱 |
| username | VARCHAR(64) UNIQUE | 用户名 |
| password_hash | VARCHAR(255) | bcrypt 哈希 |
| display_name | VARCHAR(128) | 显示名称 |
| avatar_url | TEXT | 头像 URL |
| role | ENUM(user, admin) | 角色 |
| status | ENUM(active, suspended) | 状态 |
| created_at | TIMESTAMPTZ | 创建时间 |
| updated_at | TIMESTAMPTZ | 更新时间 |

### 3.2 机器人表 (bots)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| user_id | CHAR(26) FK | 所属用户 |
| name | VARCHAR(128) | 机器人名称 |
| description | TEXT | 描述 |
| framework | ENUM(astrbot, nonebot, wechaty, koishi, custom) | 框架类型 |
| status | ENUM(online, offline, error) | 在线状态 |
| access_token | VARCHAR(64) UNIQUE | 接入令牌 |
| webhook_url | TEXT | 回调地址 |
| config | JSONB | 框架特定配置 |
| last_active_at | TIMESTAMPTZ | 最后活跃时间 |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

**索引**: user_id, access_token

### 3.3 资产表 (assets)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| user_id | CHAR(26) FK | 所属用户 |
| name | VARCHAR(255) | 资产名称 |
| description | TEXT | 描述 |
| category | ENUM(model_3d, model_live2d, motion, texture) | 分类 |
| format | VARCHAR(16) | 文件格式(vrm, fbx, glb, moc3, bvh, vmd) |
| file_key | VARCHAR(512) | MinIO 对象 key |
| file_size | BIGINT | 文件大小(字节) |
| thumbnail_key | VARCHAR(512) | 缩略图 key |
| metadata | JSONB | 模型元数据(骨骼、动画列表、时长等) |
| tags | TEXT[] | 标签数组 |
| is_public | BOOLEAN DEFAULT false | 是否公开 |
| download_count | INT DEFAULT 0 | 下载计数 |
| status | ENUM(processing, ready, failed) | 处理状态 |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

**索引**: user_id, category, format, is_public, tags (GIN)

### 3.4 虚拟形象配置表 (avatar_configs)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| user_id | CHAR(26) FK | 所属用户 |
| bot_id | CHAR(26) FK NULL | 绑定的机器人 |
| name | VARCHAR(128) | 形象名称 |
| description | TEXT | 描述 |
| render_type | ENUM(three_d, live2d) | 渲染类型 |
| scene_config | JSONB | 场景参数(相机/灯光/背景) |
| action_mapping | JSONB | 动作映射规则 |
| is_default | BOOLEAN DEFAULT false | 是否默认形象 |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

**索引**: user_id, bot_id

### 3.5 虚拟形象-资产关联表 (avatar_assets)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| avatar_id | CHAR(26) FK CASCADE | 虚拟形象 |
| asset_id | CHAR(26) FK | 资产 |
| role | ENUM(primary_model, animation, texture, accessory) | 资产角色 |
| config | JSONB | 绑定配置(骨骼映射/缩放等) |
| sort_order | INT DEFAULT 0 | 排序 |

**唯一约束**: (avatar_id, asset_id)

### 3.6 消息日志表 (message_logs)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| bot_id | CHAR(26) FK | 机器人 |
| direction | ENUM(inbound, outbound) | 消息方向 |
| content | TEXT | 消息内容 |
| metadata | JSONB | 原始消息元数据 |
| action_triggered | VARCHAR(64) | 触发的动作名 |
| created_at | TIMESTAMPTZ | |

**索引**: (bot_id, created_at)

### 3.7 动作映射模板表 (action_templates)
| 字段 | 类型 | 说明 |
|------|------|------|
| id | CHAR(26) | 主键 |
| name | VARCHAR(128) | 模板名称 |
| description | TEXT | 描述 |
| trigger_type | ENUM(keyword, emotion, regex, event, custom) | 触发类型 |
| trigger_config | JSONB | 触发条件配置 |
| action_config | JSONB | 动作执行配置 |
| is_system | BOOLEAN DEFAULT false | 是否系统内置 |
| created_at | TIMESTAMPTZ | |

---

## 4. API 设计

### 4.1 认证接口 `/api/v1/auth`
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /register | 用户注册 |
| POST | /login | 用户登录, 返回 JWT |
| POST | /refresh | 刷新 Token |
| POST | /logout | 登出(Token 加入黑名单) |

### 4.2 用户接口 `/api/v1/users`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /me | 获取当前用户信息 |
| PUT | /me | 更新用户信息 |
| PUT | /me/password | 修改密码 |

### 4.3 机器人接口 `/api/v1/bots`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | / | 获取用户的机器人列表 |
| POST | / | 创建机器人 |
| GET | /:id | 获取机器人详情 |
| PUT | /:id | 更新机器人配置 |
| DELETE | /:id | 删除机器人 |
| POST | /:id/regenerate-token | 重新生成访问令牌 |
| GET | /:id/status | 获取机器人在线状态 |

### 4.4 机器人网关接口 `/api/v1/gateway`
| 方法 | 路径 | 说明 |
|------|------|------|
| POST | /webhook/:token | 通用 Webhook 接收(Bot → 平台) |
| POST | /message | Bot 推送消息(Bearer Token 认证) |
| GET | /ws | WebSocket 长连接接入点(Bot 或前端) |

### 4.5 资产接口 `/api/v1/assets`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | / | 资产列表(分页/筛选/搜索) |
| POST | /upload | 上传资产文件(小文件 multipart) |
| POST | /upload/presigned | 获取预签名上传 URL(大文件直传 MinIO) |
| POST | /upload/complete | 确认上传完成, 触发后处理 |
| GET | /:id | 获取资产详情 |
| PUT | /:id | 更新资产元数据 |
| DELETE | /:id | 删除资产 |
| GET | /:id/download | 获取预签名下载 URL |
| GET | /:id/thumbnail | 获取缩略图 URL |
| GET | /public | 浏览公开资产库 |

### 4.6 虚拟形象接口 `/api/v1/avatars`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | / | 获取用户的虚拟形象列表 |
| POST | / | 创建虚拟形象配置 |
| GET | /:id | 获取虚拟形象详情(含关联资产) |
| PUT | /:id | 更新虚拟形象配置 |
| DELETE | /:id | 删除虚拟形象 |
| POST | /:id/bind-bot | 绑定机器人 |
| POST | /:id/bind-asset | 绑定资产 |
| DELETE | /:id/assets/:assetId | 解绑资产 |
| PUT | /:id/action-mapping | 更新动作映射规则 |

### 4.7 交互接口 `/api/v1/interaction`
| 方法 | 路径 | 说明 |
|------|------|------|
| GET | /templates | 获取动作映射模板列表 |
| POST | /templates | 创建自定义动作模板 |
| POST | /preview | 预览动作效果(发送测试指令) |

---

## 5. WebSocket 事件协议

```typescript
// 客户端 -> 服务端
interface ClientEvents {
  "avatar:join": { avatarId: string };      // 加入形象房间
  "avatar:leave": { avatarId: string };     // 离开形象房间
  "bot:test-message": {                     // 测试消息(调试用)
    botId: string;
    content: string;
  };
}

// 服务端 -> 客户端
interface ServerEvents {
  "bot:message": {                          // Bot 消息到达
    botId: string;
    content: string;
    timestamp: number;
  };
  "avatar:action": AvatarActionCommand;     // 动作指令
  "avatar:state": {                         // 状态同步
    avatarId: string;
    state: string;
  };
  "bot:status": {                           // Bot 状态变更
    botId: string;
    status: "online" | "offline" | "error";
  };
  "error": { code: string; message: string };
}

// 动作指令结构
interface AvatarActionCommand {
  avatarId: string;
  type: "animation" | "expression" | "lipsync" | "look_at" | "compound";
  action: string;          // idle, talk, happy, sad, wave 等
  params?: {
    duration?: number;     // 持续时间(ms)
    blend?: number;        // 混合权重 0-1
    loop?: boolean;
    text?: string;         // lip-sync 文本
    audioUrl?: string;     // lip-sync 音频
    target?: { x: number; y: number; z: number };
  };
  priority: number;        // 优先级(越大越优先)
  timestamp: number;
}
```

---

## 6. 核心设计模式

### 6.1 适配器模式 — 机器人接入

```go
// backend/internal/adapter/adapter.go
type BotAdapter interface {
    Framework() string
    ParseMessage(rawPayload []byte) (*BotMessage, error)
    ValidateWebhook(r *http.Request) error
    SendMessage(ctx context.Context, botID string, msg *OutboundMessage) error
}

// 每个框架实现该接口
type AstrBotAdapter struct { ... }
type NoneBotAdapter struct { ... }
type WebhookAdapter struct { ... }  // 通用 HTTP Webhook
```

### 6.2 工厂模式 — 适配器/加载器创建

```go
// backend/internal/adapter/factory.go
type AdapterFactory struct {
    adapters map[string]BotAdapter
}
func (f *AdapterFactory) Get(framework string) (BotAdapter, error)
func (f *AdapterFactory) Register(adapter BotAdapter)
```

```typescript
// frontend/src/engine/loaders/LoaderFactory.ts
class LoaderFactory {
  static create(format: ModelFormat): IModelLoader
}
```

### 6.3 策略模式 — 动作映射

```go
// backend/internal/interaction/strategy/strategy.go
type ActionMappingStrategy interface {
    Match(message *BotMessage, ctx *InteractionContext) bool
    Map(message *BotMessage, ctx *InteractionContext) []*AvatarActionCommand
    Priority() int
}

// 实现: KeywordStrategy, EmotionStrategy, RegexStrategy, CustomRuleStrategy
```

### 6.4 观察者模式 — 事件总线

```go
// backend/internal/event/bus.go
type EventBus interface {
    Publish(ctx context.Context, event Event) error
    Subscribe(eventType string, handler EventHandler) error
    Unsubscribe(eventType string, handler EventHandler) error
}

// 事件类型:
// "bot.message.received"   - Bot 消息到达
// "bot.status.changed"     - Bot 状态变更
// "avatar.action.trigger"  - 动作触发
// "asset.upload.completed"  - 资产上传完成
// "asset.process.completed" - 资产处理完成
```

### 6.5 状态机模式 — 动画控制 (前端)

```typescript
// frontend/src/engine/animation/StateMachine.ts
enum AnimationState { IDLE, TALKING, EMOTION, CUSTOM, TRANSITION }

class AnimationStateMachine {
  private currentState: AnimationState;
  private transitions: Map<string, StateTransition>;

  transition(command: AvatarActionCommand): void;
  update(deltaTime: number): void;
  getCurrentAnimation(): AnimationAction | null;
}
```

### 6.6 仓储模式 — 数据访问 (Go)

```go
// backend/internal/repository/bot_repo.go
type BotRepository interface {
    Create(ctx context.Context, bot *model.Bot) error
    FindByID(ctx context.Context, id string) (*model.Bot, error)
    FindByUserID(ctx context.Context, userID string, opts ...QueryOption) ([]*model.Bot, error)
    FindByAccessToken(ctx context.Context, token string) (*model.Bot, error)
    Update(ctx context.Context, bot *model.Bot) error
    Delete(ctx context.Context, id string) error
}
```

---

## 7. 核心模块详细设计

### 7.1 机器人适配器模块

**消息流转**:
```
Bot Framework --> HTTP Webhook/WS --> AdapterFactory.Get(framework)
  --> adapter.ParseMessage() --> EventBus.Publish("bot.message.received")
  --> InteractionService.HandleMessage() --> ActionMapper.Map()
  --> WebSocket Hub.Broadcast(avatarId, actionCommands)
  --> Frontend StateMachine --> 3D/Live2D Render
```

**AstrBot 适配方案**: AstrBot 侧编写一个轻量级转发插件,在 `@filter.event_message_type(ALL)` 中将消息 POST 到 UBotHub `/api/v1/gateway/webhook/:token`。消息格式:
```json
{
  "type": "message",
  "content": "你好",
  "sender": { "id": "user123", "name": "Alice" },
  "group": { "id": "group456", "name": "测试群" },
  "platform": "qq",
  "timestamp": 1709550000
}
```

### 7.2 资产管理模块

**上传流程**:
1. 前端请求预签名 URL → 后端通过 MinIO SDK 生成 → 返回 presigned PUT URL
2. 前端直传文件到 MinIO(绕过后端, 支持大文件)
3. 前端调用 `/upload/complete` 通知后端
4. 后端创建 Asset 记录(status=processing), 投递 asynq 任务
5. asynq Worker: 校验格式 → 提取元数据 → 生成缩略图 → 更新 status=ready
6. 通过 EventBus 通知前端资产就绪

**元数据提取**: 使用 Go 库解析模型文件头部信息:
- glTF/GLB: 解析 JSON chunk 获取 mesh/animation/material 列表
- VRM: 解析 glTF extension 获取 humanoid bone mapping
- FBX: 读取 FBX 头部获取基本信息
- Live2D: 解析 model3.json 获取 motion/expression 列表

### 7.3 渲染引擎模块 (前端)

**架构**: 基于 @react-three/fiber 的声明式 3D 渲染 + pixi-live2d-display 的 Live2D 渲染

**3D 渲染栈**:
```
<Canvas> (R3F)
  ├── <CameraRig />         # 相机控制
  ├── <Environment />        # 环境光照 (drei)
  ├── <VRMAvatar />          # VRM 角色 (或 GLTFAvatar / FBXAvatar)
  │     ├── useModelLoader() # 加载模型
  │     ├── useAnimation()   # 动画控制
  │     └── useLipSync()     # 口型同步
  └── <OrbitControls />      # 用户交互控制
```

**渲染优化**:
- 模型加载在 Web Worker 中执行(three-stdlib 的 Worker loaders)
- LOD: drei 的 `<Detailed>` 组件, 按距离切换精细度
- 纹理: KTX2Loader 压缩纹理(需 basis_transcoder WASM)
- 帧率控制: R3F `frameloop="demand"` 按需渲染, 减少空闲时 GPU 消耗

### 7.4 交互引擎模块

**动作映射流水线**:
```
BotMessage
  --> StrategyChain.Process()
    --> KeywordStrategy.Match() ?
    --> EmotionStrategy.Match() ?   // 关键词情绪推断
    --> RegexStrategy.Match() ?
    --> DefaultStrategy (fallback to idle/talk)
  --> []*AvatarActionCommand (按 priority 排序)
  --> WebSocket Broadcast
  --> 前端 StateMachine.transition()
```

**内置动作类型**:
| 动作 | 说明 | 3D 实现 | Live2D 实现 |
|------|------|---------|------------|
| idle | 待机呼吸 | AnimationClip 循环 | 自动呼吸参数 |
| talk | 说话 + 口型 | BlendShape/骨骼动画 + viseme | ParamMouthOpenY |
| happy | 高兴 | 表情 BlendShape + 动作 | Expression preset |
| sad | 悲伤 | 表情 BlendShape + 动作 | Expression preset |
| angry | 生气 | 表情 BlendShape + 动作 | Expression preset |
| surprise | 惊讶 | 表情 BlendShape + 动作 | Expression preset |
| wave | 挥手 | 骨骼动画 | Motion 文件 |
| nod | 点头 | 骨骼动画 | ParamAngleY |
| custom:* | 自定义 | 用户上传动画 | 用户上传 motion |

**口型同步方案**:
- 文本驱动(默认): 基于简易音素映射, 将文本拆分为音素序列, 对应 viseme 权重时间线
- 音频驱动(进阶): 通过 Web Audio API 获取实时频谱/RMS, 映射为口部参数(参考 ZerolanLiveRobot WaveHandler)

---

## 8. 中间件、缓存与消息队列设计

### 8.1 Gin 中间件链设计

请求处理采用洋葱模型, 中间件按顺序执行:

```
Request
  --> Recovery (panic 恢复, 防止单请求崩溃影响全局)
  --> RequestID (为每个请求生成唯一 trace_id, 贯穿日志链路)
  --> Logger (请求日志: method, path, status, latency, trace_id)
  --> CORS (跨域配置)
  --> SecurityHeaders (X-Content-Type-Options, X-Frame-Options 等)
  --> RateLimiter (滑动窗口限流, Redis 计数)
  --> [需认证路由] JWTAuth (Token 解析 + 黑名单校验)
  --> [需授权路由] RBAC (角色权限校验)
  --> Handler (业务处理)
```

**中间件实现文件**:
```
backend/internal/middleware/
  ├── recovery.go          # Panic 恢复 + 错误上报
  ├── request_id.go        # X-Request-ID 生成与传播
  ├── logger.go            # 结构化请求日志(zap)
  ├── cors.go              # CORS 配置(按环境区分)
  ├── security_headers.go  # HTTP 安全响应头
  ├── rate_limiter.go      # Redis 滑动窗口限流
  ├── jwt_auth.go          # JWT 认证
  └── rbac.go              # 基于角色的访问控制
```

**RequestID 链路追踪**: 每个请求生成 `X-Request-ID`, 写入 context, 所有日志/下游调用/错误响应均携带此 ID, 便于问题排查。

### 8.2 Redis 缓存设计

**缓存架构**: 采用 Cache-Aside (旁路缓存) 模式, 读时先查缓存, 未命中则查库并回填; 写时先更新数据库, 再删除缓存(而非更新, 防止并发问题)。

**缓存层级与 Key 设计**:

| 缓存项 | Key 格式 | TTL | 说明 |
|--------|---------|-----|------|
| 用户信息 | `user:{id}` | 30 min | 减少高频读用户表 |
| Bot 配置 | `bot:{id}` | 15 min | Bot 详情缓存 |
| Bot Token 映射 | `bot:token:{token}` | 60 min | Webhook 认证快速查找 |
| 资产元数据 | `asset:{id}` | 30 min | 资产详情缓存 |
| Avatar 配置 | `avatar:{id}` | 15 min | 含关联的资产和映射规则 |
| 预签名 URL | `presign:{assetId}:{userId}` | 55 min | 避免重复生成(URL 有效期 1h) |
| 动作模板列表 | `templates:system` | 24 h | 系统内置模板, 低频变更 |
| 用户配额 | `quota:{userId}` | 不过期 | 存储使用量, INCRBY 原子更新 |
| 登录失败计数 | `login:fail:{ip}` | 15 min | IP 锁定(自动过期释放) |
| JWT 黑名单 | `jwt:blacklist:{jti}` | = Token 剩余有效期 | 登出/密码修改时加入 |
| Bot 在线状态 | `bot:online:{botId}` | 60 s | 心跳续期, 过期即 offline |

**缓存工具封装** (`backend/internal/cache/`):
```go
// cache.go - 通用缓存接口
type Cache interface {
    Get(ctx context.Context, key string, dest interface{}) error
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
}

// redis_cache.go - Redis 实现
type RedisCache struct {
    client *redis.Client
}

// cached.go - 缓存装饰器(通用 cache-aside 封装)
func Cached[T any](cache Cache, key string, ttl time.Duration, loader func() (T, error)) (T, error)
```

**缓存一致性策略**:
- 写操作: 先写 DB, 再 Delete Cache(延迟双删可选, 对本项目写频率不高, 简单删除即可)
- Bot 状态: 使用 Redis SETEX 做心跳, 过期自动判定 offline(无需主动删除)
- 列表查询: 不缓存列表(变化频繁, 一致性难保证), 仅缓存单条记录
- 缓存穿透防护: 对 DB 查询结果为空的也缓存短 TTL 空值(5 min), 防止恶意查询

**分布式锁** (用于资产处理去重):
```go
// lock.go - Redis 分布式锁
type DistributedLock interface {
    Acquire(ctx context.Context, key string, ttl time.Duration) (bool, error)
    Release(ctx context.Context, key string) error
}
```
使用场景: 防止同一资产被重复处理(上传完成回调可能重试)。

### 8.3 消息队列设计 (asynq)

**架构**: 使用 asynq (基于 Redis Streams) 作为异步任务队列。相比 Kafka/RabbitMQ, asynq 更轻量, 无需额外部署中间件, 适合中小企业初期规模。后续若消息量增长到需要 Kafka 级别, 可通过接口抽象平滑迁移。

**队列拓扑**:
```
Producer (API Handler / EventBus)
    |
    v
Redis Streams (asynq broker)
    |
    +--> Queue: "critical"  (高优先级, 并发 10)
    |     ├── MessageDispatchTask    # Bot 消息分发到 WebSocket
    |     └── BotStatusChangeTask    # Bot 状态变更通知
    |
    +--> Queue: "default"   (普通优先级, 并发 5)
    |     ├── AssetProcessTask       # 资产格式校验 + 元数据提取
    |     ├── ThumbnailGenTask       # 缩略图生成
    |     └── ActionMappingTask      # 复杂动作映射计算
    |
    +--> Queue: "low"       (低优先级, 并发 2)
          ├── AssetCleanupTask       # 过期/删除资产清理
          ├── MessageLogArchiveTask  # 历史消息归档
          └── MetricsAggregateTask   # 指标聚合统计
```

**任务定义**:
```go
// backend/internal/queue/tasks/

// asset_process.go - 资产处理任务
type AssetProcessPayload struct {
    AssetID  string `json:"asset_id"`
    UserID   string `json:"user_id"`
    FileKey  string `json:"file_key"`
    Format   string `json:"format"`
}

// message_dispatch.go - 消息分发任务
type MessageDispatchPayload struct {
    BotID     string `json:"bot_id"`
    AvatarID  string `json:"avatar_id"`
    Content   string `json:"content"`
    Metadata  map[string]interface{} `json:"metadata"`
    Timestamp int64  `json:"timestamp"`
}

// thumbnail_gen.go - 缩略图生成
type ThumbnailGenPayload struct {
    AssetID string `json:"asset_id"`
    FileKey string `json:"file_key"`
    Format  string `json:"format"`
}
```

**任务可靠性**:
- 自动重试: 失败任务最多重试 3 次, 指数退避(1s, 4s, 16s)
- 死信队列: 超过重试次数的任务进入 "archived" 状态, 可在管理面板查看
- 唯一性: 资产处理任务使用 `asynq.Unique(1h)` 防止重复投递
- 超时: 资产处理 5 min, 缩略图生成 2 min, 消息分发 10s
- 监控: asynq 内置 Web UI (asynqmon) 可查看队列状态、任务详情、失败率

**asynq 管理面板**: 开发环境启动 `asynqmon` 服务(端口 8081), 可视化管理任务队列:
```yaml
# docker-compose.yml
asynqmon:
  image: hibiken/asynqmon
  ports:
    - "8081:8080"
  environment:
    - REDIS_ADDR=redis:6379
```

### 8.4 对象存储抽象层设计

**Strategy 模式**: 通过统一接口抽象存储后端, 支持本地 MinIO 和阿里云 OSS 无缝切换。

```go
// backend/internal/service/storage/

// storage.go - 统一接口
type ObjectStorage interface {
    // 上传文件
    PutObject(ctx context.Context, bucket, key string, reader io.Reader, size int64, contentType string) error
    // 获取文件
    GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, error)
    // 删除文件
    DeleteObject(ctx context.Context, bucket, key string) error
    // 生成预签名上传 URL
    PresignedPutURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
    // 生成预签名下载 URL
    PresignedGetURL(ctx context.Context, bucket, key string, expires time.Duration) (string, error)
    // 检查文件是否存在
    ObjectExists(ctx context.Context, bucket, key string) (bool, error)
    // 获取文件元信息
    StatObject(ctx context.Context, bucket, key string) (*ObjectInfo, error)
}

// minio_storage.go - MinIO 实现(开发环境/私有部署)
type MinIOStorage struct {
    client *minio.Client
}

// aliyun_oss_storage.go - 阿里云 OSS 实现(生产环境)
type AliyunOSSStorage struct {
    client *oss.Client
    bucket *oss.Bucket
}
```

**配置切换**:
```yaml
# config.yaml
storage:
  provider: "minio"  # minio | aliyun_oss
  minio:
    endpoint: "localhost:9000"
    access_key: "minioadmin"
    secret_key: "minioadmin"
    use_ssl: false
    bucket: "ubothub-assets"
  aliyun_oss:
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    access_key_id: "${ALIYUN_ACCESS_KEY_ID}"
    access_key_secret: "${ALIYUN_ACCESS_KEY_SECRET}"
    bucket: "ubothub-assets"
    cdn_domain: "cdn.ubothub.com"  # CDN 加速域名(可选)
```

**CDN 加速**(阿里云 OSS 生产环境):
- 模型文件下载通过 CDN 域名分发, 减少 OSS 直连流量费用
- 预签名 URL 使用 CDN 域名替换 OSS 域名
- CDN 缓存策略: 模型文件缓存 7 天(按 file_key 包含 assetId, 更新即新 key)

**阿里云 OSS 依赖**: `github.com/aliyun/aliyun-oss-go-sdk/oss`

### 8.5 事件总线详细设计

**实现**: 基于 Go channel 的进程内事件总线(单体应用阶段)。若未来拆分微服务, 可替换为 Redis Pub/Sub 或 NATS。

```go
// backend/internal/event/

// bus.go
type EventBus struct {
    subscribers map[string][]EventHandler
    mu          sync.RWMutex
    workerPool  chan struct{}  // 控制并发处理数
}

// 异步发布: 事件投递后立即返回, 由 goroutine 池异步处理
func (b *EventBus) Publish(ctx context.Context, event Event) error

// 同步发布: 等待所有 handler 执行完毕(用于关键流程)
func (b *EventBus) PublishSync(ctx context.Context, event Event) error
```

**事件流转示意**:
```
[Bot Webhook 到达]
  --> AdapterService.HandleWebhook()
  --> EventBus.Publish("bot.message.received", BotMessageEvent)
      |
      +--> InteractionService.OnBotMessage()
      |     --> ActionMapper.Map()
      |     --> WebSocketHub.Broadcast(avatarId, commands)
      |
      +--> MessageLogService.OnBotMessage()
            --> 异步写入 message_logs 表

[资产上传完成]
  --> AssetHandler.CompleteUpload()
  --> EventBus.Publish("asset.upload.completed", AssetEvent)
      |
      +--> QueueProducer.EnqueueAssetProcess()
            --> asynq 投递处理任务
```

### 8.6 目录结构补充

上述新增模块的目录:
```
backend/internal/
  ├── cache/                    # 缓存层
  │   ├── cache.go              # Cache 接口定义
  │   ├── redis_cache.go        # Redis Cache 实现
  │   ├── cached.go             # Cache-Aside 装饰器
  │   └── lock.go               # 分布式锁
  ├── service/
  │   └── storage/              # 对象存储层
  │       ├── storage.go        # ObjectStorage 接口
  │       ├── minio_storage.go  # MinIO 实现
  │       └── aliyun_oss_storage.go  # 阿里云 OSS 实现
  ├── middleware/                # (已有, 补充完善)
  ├── event/                    # (已有, 补充完善)
  └── queue/                    # (已有, 补充完善)
```

---

## 9. 开发路线图

### Phase 1: 基础框架搭建 (2-3 周)
- [ ] 初始化项目: Go module + React + Docker Compose
- [ ] Go 后端骨架: Gin + GORM + Viper + Wire + Zap
- [ ] 中间件链: Recovery, RequestID, Logger, CORS, SecurityHeaders, RateLimiter
- [ ] Redis 缓存层: Cache 接口 + Redis 实现 + Cache-Aside 装饰器
- [ ] React 前端骨架: Vite + Ant Design + Router + Zustand
- [ ] Docker Compose: PostgreSQL + Redis + MinIO + asynqmon
- [ ] 用户认证模块: 注册/登录/JWT/黑名单
- [ ] 对象存储抽象层: MinIO + 阿里云 OSS 双实现
- [ ] 统一响应格式 + 错误码体系
- [ ] Swagger API 文档配置
- [ ] ESLint + Prettier + Makefile

### Phase 2: 机器人管理 (2 周)
- [ ] Bot CRUD API + Repository + Service + Handler
- [ ] BotAdapter 接口定义 + Factory
- [ ] 通用 Webhook 适配器实现
- [ ] AstrBot 适配器实现
- [ ] Bot 心跳检测与状态管理(Redis SETEX)
- [ ] WebSocket Hub 基础实现(gorilla/websocket)
- [ ] 事件总线(EventBus)
- [ ] asynq 任务队列 + Worker 框架
- [ ] 前端: Bot 管理页面(列表/创建/配置)

### Phase 3: 资产管理 (2 周)
- [ ] MinIO 存储服务封装(上传/下载/预签名)
- [ ] 文件上传 API(直传 + 预签名)
- [ ] asynq 任务队列配置
- [ ] 资产处理 Worker(格式校验 + 元数据提取)
- [ ] 缩略图生成
- [ ] 前端: 资产库页面(上传/浏览/管理/搜索)

### Phase 4: 3D 渲染引擎 (3 周)
- [ ] R3F 场景容器 + 环境配置
- [ ] VRM 模型加载与渲染(VRMAvatar)
- [ ] GLTF/GLB 模型加载
- [ ] FBX 模型加载
- [ ] 动画控制器 + AnimationMixer 封装
- [ ] 动画状态机(StateMachine)
- [ ] 模型加载器工厂(LoaderFactory)
- [ ] 前端: 模型预览组件

### Phase 5: Live2D 渲染引擎 (2 周)
- [ ] PixiJS + pixi-live2d-display 集成
- [ ] Live2D 模型加载(moc3 + model3.json)
- [ ] Live2D 参数控制(表情/动作)
- [ ] Live2D 口型同步
- [ ] 前端: Live2D 预览组件
- [ ] 3D/Live2D 统一切换接口

### Phase 6: 交互联动 (2-3 周)
- [ ] InteractionService(消息 -> 动作映射)
- [ ] ActionMapper + Strategy Chain
- [ ] 策略实现: Keyword, Emotion, Regex, Default
- [ ] AvatarConfig 管理(绑定 Bot + Model + Actions)
- [ ] 全链路打通: Bot 消息 -> 后端 -> WebSocket -> 前端动画
- [ ] 口型同步(文本驱动 + 音频驱动)
- [ ] 前端: 虚拟形象编辑器 + 实时交互页面

### Phase 7: 完善与优化 (2 周)
- [ ] 渲染优化: Web Worker 加载, LOD, 纹理压缩
- [ ] NoneBot/Wechaty/Koishi 适配器补充
- [ ] 前端 UI 完善 + 响应式
- [ ] Prometheus metrics 暴露 + Grafana 配置
- [ ] 单元测试 + 集成测试
- [ ] Docker 生产镜像 + 部署文档
- [ ] projdesign.md 最终整理

---

## 10. 安全设计

### 10.1 认证与授权安全

**JWT 安全策略**:
- Access Token 有效期 15 分钟, Refresh Token 有效期 7 天
- Refresh Token 存储于 Redis, 支持主动吊销(logout 时删除)
- Token 黑名单机制: logout/密码修改后将旧 Token 加入 Redis 黑名单(TTL = Token 剩余有效期)
- JWT 签名算法使用 HS256(对称)或 RS256(非对称, 适合微服务场景)
- 禁止在 JWT payload 中存放敏感信息(密码、完整邮箱等)

**密码安全**:
- 使用 bcrypt 哈希(cost factor >= 12)
- 密码强度校验: 最少 8 字符, 包含大小写字母 + 数字
- 登录失败锁定: 同一 IP 连续 5 次失败后锁定 15 分钟(Redis 计数)
- 密码重置 Token 一次性使用, 有效期 30 分钟

**RBAC 权限模型**:
- 角色: admin(平台管理员), user(普通用户)
- 资源级鉴权: 用户只能操作自己的 Bot/Asset/Avatar(Handler 层校验 user_id)
- Bot access_token 使用 crypto/rand 生成 32 字节随机串, 不可预测

### 10.2 API 安全

**输入校验与防注入**:
- 所有请求体通过 go-playground/validator 进行结构体校验(required, min, max, email 等)
- SQL 注入防护: GORM 参数化查询(禁止拼接 SQL)
- 路径参数使用正则白名单校验(如 ID 格式 `^[a-z0-9]{20,30}$`)
- 禁止直接将用户输入写入日志(防止日志注入)

**请求限流(Rate Limiting)**:
- 基于 Redis 的滑动窗口限流中间件
- 全局: 100 req/s per IP
- 认证接口: 10 req/min per IP(防暴力破解)
- 文件上传: 10 req/min per User
- WebSocket 消息: 60 msg/min per Connection
- Bot Webhook: 300 req/min per Token

**CORS 配置**:
- 生产环境严格限制 Allow-Origin 为前端域名
- 禁止 `Access-Control-Allow-Origin: *`
- 明确指定允许的 Methods 和 Headers

**HTTP 安全头**(通过 Gin 中间件统一设置):
```
X-Content-Type-Options: nosniff
X-Frame-Options: DENY
X-XSS-Protection: 1; mode=block
Strict-Transport-Security: max-age=31536000; includeSubDomains
Content-Security-Policy: default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'
Referrer-Policy: strict-origin-when-cross-origin
Permissions-Policy: camera=(), microphone=(), geolocation=()
```

**API 版本控制**: 所有接口以 `/api/v1/` 为前缀, 便于后续不兼容升级

### 10.3 WebSocket 安全

- 连接时验证 JWT(通过 query param 或首条消息携带 token)
- 连接后绑定 user_id, 只推送该用户有权限的事件
- 心跳检测: 30 秒无消息则服务端发送 ping, 客户端 10 秒内未 pong 则断开
- 消息体大小限制: 单条消息最大 64KB
- 消息频率限制: 单连接 60 msg/min, 超限断开并返回错误码
- 广播隔离: 通过 room 机制(avatarId)隔离不同用户的消息流

### 10.4 文件上传安全

**上传校验**:
- 文件类型白名单: 仅允许 `.vrm`, `.fbx`, `.glb`, `.gltf`, `.zip`(Live2D 打包), `.bvh`, `.vmd`
- Magic Bytes 检测: 不信任文件扩展名, 读取文件头部字节验证实际格式
  - glTF Binary: `0x676C5446` ("glTF")
  - FBX Binary: `Kaydara FBX Binary`
  - ZIP: `0x504B0304`
- 文件大小限制: 单文件最大 500MB, 用户总存储配额 5GB(可配置)
- 文件名消毒: 移除特殊字符, 使用 UUID 重命名存储, 原始文件名仅存于数据库元数据

**存储安全**:
- MinIO Bucket 策略: 私有访问, 通过预签名 URL 临时授权(有效期 1 小时)
- 预签名 URL 仅在已认证用户请求时生成, 绑定特定对象 key
- 禁止目录遍历: 对象 key 使用 `{userId}/{assetId}/{filename}` 结构
- 上传完成后的异步处理在隔离环境中执行(防止恶意文件影响主进程)

**防病毒/恶意文件**:
- 异步 Worker 处理时校验文件结构完整性(解析失败则标记为 FAILED)
- Live2D ZIP 包: 解压前检测路径穿越(ZipSlip 攻击), 限制解压后总大小

### 10.5 前端安全

**XSS 防护**:
- React 默认对 JSX 表达式进行 HTML 转义
- 禁止使用 `dangerouslySetInnerHTML`(除非经过 DOMPurify 消毒)
- Bot 消息内容在展示前进行转义处理
- CSP 头限制脚本来源

**CSRF 防护**:
- API 采用 JWT Bearer Token 认证(非 Cookie), 天然免疫 CSRF
- 如后续引入 Cookie Session, 需添加 CSRF Token 机制

**敏感信息保护**:
- access_token 在前端仅展示一次(创建时), 之后不可再查看
- JWT 存储于内存(Zustand store), 不写入 localStorage(防 XSS 窃取)
- Refresh Token 存于 httpOnly Cookie(如采用 Cookie 方案)

**3D/Live2D 渲染安全**:
- 模型文件从可信来源加载(MinIO 预签名 URL)
- WebGL 上下文: 监控 context lost 事件, 防止 GPU 资源耗尽攻击
- Web Worker 加载模型: 隔离解析过程, 防止恶意模型文件导致主线程崩溃

### 10.6 基础设施安全

**数据库安全**:
- PostgreSQL 连接使用 SSL/TLS
- 数据库账户最小权限原则(应用账户仅 CRUD, 禁止 DROP/ALTER)
- 敏感字段加密存储: password_hash(bcrypt), 考虑对 access_token 加密存储

**Redis 安全**:
- 设置 requirepass 密码认证
- 绑定内网 IP, 禁止公网暴露
- 禁用危险命令: FLUSHALL, FLUSHDB, KEYS, CONFIG

**MinIO 安全**:
- 访问密钥定期轮换
- Bucket 策略: 禁止匿名访问
- 开启访问日志审计

**配置与密钥管理**:
- 敏感配置(数据库密码、JWT Secret、MinIO 密钥)通过环境变量注入, 不写入代码仓库
- `.env` 文件加入 `.gitignore`
- 生产环境考虑使用 HashiCorp Vault 或云 KMS 管理密钥
- JWT Secret 最少 256 bit 随机生成

**Docker 安全**:
- 使用非 root 用户运行容器
- 镜像基于 distroless 或 alpine 最小化攻击面
- 固定基础镜像版本标签(不使用 `latest`)
- 定期扫描镜像漏洞(Trivy/Snyk)

### 10.7 日志与审计

**安全审计日志**(独立于业务日志):
- 用户登录/登出事件(含 IP、User-Agent)
- 密码修改/重置事件
- Bot Token 创建/重新生成
- 资产上传/删除操作
- 异常登录检测(新 IP/新设备)

**日志安全**:
- 日志中脱敏处理: 密码、Token、邮箱(部分遮掩)
- 结构化日志(JSON 格式, 便于 ELK/Loki 采集)
- 日志保留策略: 安全日志至少保留 90 天

### 10.8 安全相关依赖

```
github.com/golang-jwt/jwt/v5           # JWT
golang.org/x/crypto/bcrypt             # 密码哈希
github.com/ulule/limiter/v3            # 限流中间件
github.com/unrolled/secure             # HTTP 安全头中间件
github.com/rs/cors                     # CORS (Gin 也有内置)
```

### 10.9 安全开发规范

- 所有外部输入(HTTP Body/Query/Path/Header, WebSocket Message, Webhook Payload)必须经过校验后才能使用
- GORM 查询禁止使用 `db.Raw()` 拼接用户输入, 必须使用参数化 `db.Where("id = ?", id)`
- 错误响应不暴露内部实现细节(数据库错误信息、堆栈跟踪等), 仅返回通用错误码
- 定期执行依赖安全扫描: `govulncheck`(Go), `pnpm audit`(前端)
- Code Review 中安全检查清单: 输入校验、权限检查、SQL 注入、XSS、敏感信息泄露

---

## 11. 分布式部署预留设计

### 11.1 架构分层与拆分预留

当前采用**模块化单体**架构, 但所有核心模块通过接口隔离, 确保未来可平滑拆分为微服务:

```
Phase 1 (当前): 模块化单体
  ┌──────────────────────────────────────────────────┐
  │                   单个 Go 进程                     │
  │  ┌─────────┐ ┌──────────┐ ┌────────────────────┐ │
  │  │ Auth    │ │ Bot      │ │ Asset              │ │
  │  │ Module  │ │ Module   │ │ Module             │ │
  │  ├─────────┤ ├──────────┤ ├────────────────────┤ │
  │  │ Avatar  │ │ Gateway  │ │ Interaction        │ │
  │  │ Module  │ │ Module   │ │ Module             │ │
  │  └─────────┘ └──────────┘ └────────────────────┘ │
  └──────────────────────────────────────────────────┘

Phase 2 (未来): 微服务拆分
  ┌───────────┐  ┌───────────┐  ┌───────────────┐
  │ API 网关   │  │ WS 网关    │  │ Worker 集群   │
  │ (Go/Gin)  │  │ (Go)      │  │ (Go/asynq)   │
  └─────┬─────┘  └─────┬─────┘  └───────┬───────┘
        │               │               │
  ┌─────▼─────┐  ┌─────▼─────┐  ┌──────▼──────┐
  │ Auth Svc  │  │ Bot Svc   │  │ Asset Svc   │
  └───────────┘  └───────────┘  └─────────────┘
```

### 11.2 接口抽象保障拆分能力

以下关键组件已通过接口抽象, 后续替换实现无需修改业务代码:

| 组件 | 当前实现 | 分布式替换 |
|------|---------|-----------|
| EventBus | Go channel(进程内) | Redis Pub/Sub 或 NATS |
| Cache | Redis(单实例) | Redis Cluster / Redis Sentinel |
| 任务队列 | asynq(单 Redis) | asynq(Redis Cluster) 或 Kafka |
| 对象存储 | MinIO(单节点) | 阿里云 OSS(已实现) / MinIO 集群 |
| 数据库 | PostgreSQL(单实例) | PostgreSQL 主从 + 读写分离 |
| WebSocket | 单进程 Hub | Redis Pub/Sub 广播 + 多节点 Hub |

### 11.3 WebSocket 多节点方案预留

单体阶段 WebSocket Hub 在单进程内管理所有连接。拆分后需要跨节点广播:

```go
// gateway/hub.go - 当前: 进程内广播
type Hub interface {
    Register(client *Client)
    Unregister(client *Client)
    BroadcastToRoom(room string, message []byte)
}

// 未来: Redis Pub/Sub 跨节点广播
// 发送方 Hub 将消息 Publish 到 Redis channel "ws:{room}"
// 所有 WS 节点 Subscribe 该 channel, 转发给本地连接
```

**当前代码中预留**:
- Hub 定义为 interface, 而非 struct, 便于替换
- Room(avatarId)管理逻辑独立, 不依赖进程状态
- 消息序列化使用 JSON, 跨进程传输无障碍

### 11.4 数据库读写分离预留

GORM 原生支持读写分离, 通过 DBResolver 插件:

```go
// 当前: 单数据源
db, _ := gorm.Open(postgres.Open(dsn))

// 未来: 读写分离(仅需修改初始化代码)
db.Use(dbresolver.Register(dbresolver.Config{
    Replicas: []gorm.Dialector{
        postgres.Open(replicaDSN1),
        postgres.Open(replicaDSN2),
    },
    Policy: dbresolver.RandomPolicy{},
}))
```

**当前代码中预留**:
- Repository 层统一使用 `*gorm.DB` 而非直接连接, 切换透明
- 数据库配置支持多 DSN(Viper 配置数组)

### 11.5 无状态设计原则

确保后端进程无本地状态, 支持水平扩展:

- **Session**: 无服务端 Session, 完全使用 JWT(无状态认证)
- **文件**: 所有文件存储在 MinIO/OSS, 不存本地磁盘
- **缓存**: 全部存 Redis, 不使用进程内缓存(或双层: 本地 LRU + Redis)
- **WebSocket 状态**: 连接映射存 Redis(哪个节点持有哪些连接)
- **定时任务**: 使用 asynq 的 Scheduler(Redis 协调, 单节点执行), 避免多节点重复执行

### 11.6 配置中心预留

当前使用 Viper 读取本地 YAML + 环境变量。未来可接入:
- **Nacos / Apollo / Consul**: Viper 支持远程配置源, 仅需添加 Provider
- **配置热更新**: Viper WatchConfig + callback, 已预留接口

### 11.7 Docker 与 K8s 部署预留

```yaml
# docker-compose.prod.yml (生产环境, 单机多容器)
services:
  api:
    build: ./backend
    replicas: 2                    # 多实例
    ports: ["8080"]
    depends_on: [postgres, redis, minio]
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/api/v1/health"]
      interval: 30s
      retries: 3

  worker:
    build: ./backend
    command: ["./server", "--mode=worker"]  # Worker 模式(只跑 asynq)
    replicas: 2
    depends_on: [redis]

  frontend:
    build: ./frontend
    ports: ["3000"]

  nginx:
    image: nginx:alpine
    ports: ["80:80", "443:443"]
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    depends_on: [api, frontend]
```

**Go 二进制支持多运行模式** (通过命令行参数):
```go
// cmd/server/main.go
// --mode=api     启动 HTTP + WebSocket 服务
// --mode=worker  启动 asynq Worker
// --mode=all     启动全部(默认, 开发环境)
```

**K8s 部署清单预留** (Phase 2+):
```
deployments/k8s/
  ├── api-deployment.yaml
  ├── api-service.yaml
  ├── api-hpa.yaml           # HPA 自动扩缩容
  ├── worker-deployment.yaml
  ├── frontend-deployment.yaml
  ├── ingress.yaml
  └── configmap.yaml
```

### 11.8 可观测性(分布式追踪预留)

当前: RequestID 贯穿单进程日志链路。

未来升级路径:
- OpenTelemetry SDK (Go): 替换 RequestID 为分布式 Trace
- Jaeger / Zipkin: 追踪跨服务调用链路
- 当前 zap 日志已包含 trace_id 字段, 迁移时仅需替换 ID 生成逻辑

---

## 12. 关键依赖库清单

### Go 后端
```
github.com/gin-gonic/gin              # Web 框架
gorm.io/gorm                          # ORM
gorm.io/driver/postgres               # PostgreSQL 驱动
github.com/redis/go-redis/v9          # Redis 客户端
github.com/gorilla/websocket          # WebSocket
github.com/hibiken/asynq              # 异步任务队列
github.com/minio/minio-go/v7          # MinIO SDK
github.com/aliyun/aliyun-oss-go-sdk   # 阿里云 OSS SDK
github.com/golang-jwt/jwt/v5          # JWT
github.com/spf13/viper                # 配置管理
github.com/google/wire                # 依赖注入
github.com/swaggo/swag                # Swagger 文档
go.uber.org/zap                       # 结构化日志
github.com/go-playground/validator/v10 # 数据校验
github.com/prometheus/client_golang   # Prometheus 指标
github.com/rs/xid                     # 短 ID 生成
github.com/ulule/limiter/v3           # Redis 限流
github.com/unrolled/secure            # HTTP 安全头
golang.org/x/crypto                   # bcrypt 密码哈希
github.com/stretchr/testify           # 测试断言
github.com/golang/mock                # Mock 生成
```

### React 前端
```
react, react-dom                      # React 核心
typescript                            # TypeScript
vite, @vitejs/plugin-react            # 构建工具
antd, @ant-design/icons               # UI 组件库
three, @types/three                   # Three.js
@react-three/fiber                    # React Three.js
@react-three/drei                     # R3F 辅助
@pixiv/three-vrm                      # VRM 加载器
pixi.js, pixi-live2d-display          # Live2D
zustand                               # 状态管理
react-router-dom                      # 路由
axios                                 # HTTP 客户端
socket.io-client                      # WebSocket
react-hook-form, zod                  # 表单验证
echarts, echarts-for-react            # 图表
```

---

## 13. 验证方案

### 开发环境启动
```bash
# 1. 启动基础设施
docker-compose up -d  # PostgreSQL + Redis + MinIO

# 2. 启动后端
cd backend && air      # Go 热重载开发

# 3. 启动前端
cd frontend && pnpm dev

# 4. API 文档
open http://localhost:8080/swagger/index.html
```

### 功能验证
1. 注册/登录 → 创建 Bot → 获取 access_token
2. curl 向 `/api/v1/gateway/webhook/:token` 发送测试消息, 验证消息到达
3. 上传 VRM/GLB/Live2D 模型, 验证处理流水线(status: processing → ready)
4. 前端加载模型, 验证 3D/Live2D 渲染正常
5. 发送 Bot 消息, 验证虚拟形象实时响应(全链路)
6. 连接 AstrBot 转发插件, 验证端到端消息流转与动画联动

### 性能验证
- WebSocket 并发: 使用 k6 测试 1000+ 并发连接
- 大文件上传: 100MB+ 模型文件预签名直传
- 渲染帧率: Three.js Stats.js 面板, 目标 60fps(复杂场景 30fps+)

---

## 14. 参考模式来源

| 本项目模块 | 参考来源 | 说明 |
|-----------|---------|------|
| EventBus | ZerolanLiveRobot `event/event_emitter.py` | 适配为 Go channel-based 事件总线 |
| WebSocket Hub | ZerolanLiveRobot `services/playground/bridge.py` | 适配为 gorilla/websocket Hub 模式 |
| LipSyncController | ZerolanLiveRobot `services/live2d/wave_handler.py` | RMS 音频分析驱动口型 |
| BotAdapter Factory | AstrBot `core/factory.py` | Go interface + factory 实现 |
| Repository Pattern | AstrBot `services/database/facades/` | Go Repository 接口 + GORM 实现 |
| 3D 渲染集成 | 3d-pet `src/composable/useModel.ts` | 适配为 React hooks |
