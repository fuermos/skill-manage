# Skill Sync

多机器共享 AI 开发工具 (Claude Code / OpenCode / Trae) 的 **Skills、Rules、Agents、Memories** 同步管理平台。

## 功能特性

- 多工具支持：Claude Code、OpenCode、Trae IDE（可扩展）
- 多机器同步：服务端存储 + 客户端 push/pull
- 版本管理：revision 驱动增量同步，冲突检测
- 桌面应用：Wails + Svelte 5 原生桌面（中英文切换）
- Web UI：内嵌浏览器管理界面
- CLI：命令行 sync 操作
- 推荐引擎：协同过滤 + 标签相似度 + 调用链
- 自动记录：使用 stats / 错误案例收集
- NAS 部署：Docker Compose 一键启动

## 项目结构

```
skill-manage/
├── cmd/
│   ├── server/main.go         # 服务端入口
│   └── client/main.go         # CLI 客户端 + Web UI
├── internal/
│   ├── model/                 # 数据模型
│   ├── store/                 # SQLite 持久化
│   ├── server/                # REST API
│   ├── auth/                  # Token 认证
│   ├── adapter/               # 多工具适配器
│   ├── client/                # 同步引擎 + Web UI
│   └── engine/                # 推荐引擎
├── frontend/                  # Wails 前端 (Svelte 5)
├── configs/tools/             # 工具声明式配置
│   ├── claude.yaml
│   ├── opencode.yaml
│   └── trae.yaml
├── docker/                    # Docker 部署
├── app.go                     # Wails 绑定
├── main.go                    # Wails 入口
├── wails.json                 # Wails 配置
└── docker-compose.yml         # NAS 部署
```

## 快速开始

### 1. 启动服务端

```bash
# 直接运行
go build -o server ./cmd/server/
./server --addr :8080 --db ./server-data/sync.db --token my-secret-token

# Docker Compose (推荐 NAS 部署)
mkdir server-data
docker-compose up -d
```

### 2. 启动客户端

#### 方式 A：桌面应用（推荐）

双击 `build/bin/skill-sync.exe`，默认连接 `localhost:8080`。在 Settings 页面修改服务端地址和 Token。

#### 方式 B：Web UI

```bash
export SKILL_SYNC_SERVER=http://192.168.1.100:8080
export SKILL_SYNC_TOKEN=my-secret-token
./skill-sync ui   # 浏览器打开 localhost:3000
```

#### 方式 C：CLI

```bash
export SKILL_SYNC_SERVER=http://192.168.1.100:8080
export SKILL_SYNC_TOKEN=my-secret-token

./skill-sync status        # 查看同步状态
./skill-sync diff           # 查看本地与服务端差异
./skill-sync push -y        # 推送本地变更
./skill-sync pull           # 拉取服务端变更
./skill-sync discover trae  # 列出 Trae 本地 skills
```

## 支持的工具

| 工具 | 配置路径 | Skills 位置 |
|------|---------|------------|
| Claude Code | `~/.claude/` | `skills/*/SKILL.md` |
| OpenCode | `~/.config/opencode/` | `skills/*/SKILL.md` |
| Trae IDE | `~/.trae-cn/` | `builtin/*/skills/*/SKILL.md` |

新增工具：在 `configs/tools/` 下添加 YAML 文件，无需改代码。

## 同步机制

```
Pull:  本地 last_revision  vs  服务端 current_revision
       ├── 相等 → 无需拉取
       └── 不同 → GET /changes?since=N  增量获取

Push:  base_revision  vs  current_revision
       ├── 相等 → 直接接受 → revision +1
       └── 不同 → 冲突检测 → 返回冲突列表
```

冲突以服务端为准，本地旧版本备份到 `%APPDATA%\skill-sync\backups\`。

## API 端点

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/v1/health` | 健康检查 |
| GET | `/api/v1/sync/status` | 获取服务端 revision |
| GET | `/api/v1/sync/changes?since=N` | 增量变更列表 |
| POST | `/api/v1/sync/push` | 推送本地变更 |
| POST | `/api/v1/sync/pull` | 拉取服务端变更 |
| POST | `/api/v1/sync/files` | 批量下载文件内容 |
| GET | `/api/v1/skills?tool=&category=&search=` | 列出 skills |
| GET | `/api/v1/skills/{id}` | skill 详情 |
| POST | `/api/v1/usage/batch` | 批量上报使用记录 |
| GET | `/api/v1/combinations` | skill 组合列表 |
| GET | `/api/v1/chains` | 调用链列表 |

## 开发

### 环境要求

- Go 1.23+
- Node.js 18+ (仅 Wails 桌面端需要)
- Wails CLI: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`

### 运行测试

```bash
go test ./internal/... -count=1
```

### 开发模式

```bash
# 前端开发 (热更新)
cd frontend && npm run dev

# Wails 开发模式 (Go + 前端热更新)
wails dev

# 仅编译
wails build
```

### 添加新工具

在 `configs/tools/` 下创建 YAML：

```yaml
name: mytool
display_name: "My Tool"
enabled: true
local_path:
  windows: "%APPDATA%\\MyTool"
  linux: "$HOME/.config/mytool"
  darwin: "$HOME/.config/mytool"
server_path: "mytool"
categories:
  skills:
    local_dir: "skills"
    pattern: "**/SKILL.md"
    priority: high
ignore:
  - "**/.git/**"
on_missing: skip
```

然后在 `app.go` 和 `internal/client/ui.go` 的 `configFiles` 列表中添加该文件。

### 内网穿透

```bash
# 方案 A: frp
# 方案 B: tailscale
# 方案 C: ngrok (仅测试)
ngrok http 8080
```

## 多机器部署示意

```
┌──────────┐   ┌──────────┐   ┌──────────┐
│ Win PC   │   │ Linux PC │   │ Macbook  │
│ 桌面应用  │   │ CLI      │   │ 桌面应用  │
└─────┬─────┘   └────┬─────┘   └────┬─────┘
      │              │              │
      └──────────────┼──────────────┘
                     │ HTTP API
              ┌──────▼──────┐
              │ Server (NAS) │
              │ :8080        │
              │ SQLite       │
              └─────────────┘
```

## License

MIT