# Skill Sync 开发路线图

## 已完成 (v1.0)

| 模块 | 状态 | 说明 |
|------|:--:|------|
| 多工具适配器 | ✅ | Claude Code / OpenCode / Trae IDE, YAML 声明式配置 |
| 文件同步引擎 | ✅ | push/pull/diff, revision 增量同步, 冲突检测 |
| 服务端 REST API | ✅ | 10 endpoints, Token 认证, SQLite 存储 |
| Wails 桌面应用 | ✅ | Svelte 5, 中/EN 切换, Dashboard/Skills/Sync/Settings |
| Web UI | ✅ | Go embed, 浏览器 localhost:3000 |
| CLI 客户端 | ✅ | push/pull/diff/discover/status/config |
| SKILL.md 元数据解析 | ✅ | YAML frontmatter → name/description/summary/tags |
| Docker Compose | ✅ | NAS 一键部署 |
| 基础推荐引擎 | ✅ | 协同过滤 + 标签相似度 + 调用链 (算法完成, 待接入真实数据) |
| 数据库迁移 | ✅ | 10 表: files/skills/examples/error_cases/combos/chains/usage |

---

## 待实现 (v1.1)

### 1. 使用案例 / 成功案例 / 失败案例

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **案例提交 UI** | 🔴 高 | Skill 详情页添加"提交案例"按钮，填写 场景/输入/结果/评分 |
| **案例列表展示** | 🔴 高 | Skill 详情页展示 usage/success/failure 三 Tab 案例列表 |
| **案例统计** | 🟡 中 | Dashboard 展示最近案例、高频错误等 |
| **案例搜索** | 🟡 中 | 按错误类型/标签/skill 搜索案例 |

### 2. 错误案例自动收集分析 Skill

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **Error Case Collector Skill** | 🔴 高 | 创建 `error-case-collector` SKILL.md, 分析 session transcript |
| **会话 Transcript 解析** | 🔴 高 | 解析 Claude Code Stop hook JSONL, 识别错误-解决模式 |
| **错误聚类** | 🟡 中 | 跨 session 聚合同类错误, 计算解决率 |
| **反向增强 Skill** | 🟢 低 | 案例积累后建议更新相关 SKILL.md |

### 3. 自动收集使用记录

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **Hook 安装器** | 🔴 高 | `skill-sync hook install claude` 自动注入 Stop hook |
| **Usage Collector** | 🔴 高 | daemon 解析 transcript JSONL, 上报 skill 使用 |
| **离线缓冲** | 🟡 中 | 网络不通时本地 SQLite 暂存, 恢复后批量上报 |
| **OpenCode/Trae 适配** | 🟡 中 | 调研各自的 session/hook 机制 |

### 4. Skill 组合 (Combinations)

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **组合 CRUD UI** | 🟡 中 | 创建/编辑/删除 skill 组合, 关联 use case |
| **组合列表页** | 🟡 中 | 展示所有组合, 搜索/过滤 |
| **自动化推荐生成** | 🟢 低 | 基于使用数据自动发现常见组合 |

### 5. Skill 调用链 (Chains)

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **调用链 CRUD UI** | 🟡 中 | 创建/编辑有序步骤链 |
| **调用链可视化** | 🟢 低 | 流程图展示步骤和依赖关系 |
| **链触发推荐** | 🟢 低 | "你刚用了 A, 下一步试试 B (来自链: TDD Workflow)" |

### 6. 推荐引擎增强

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **接入真实使用数据** | 🔴 高 | 当前推荐引擎算法已有, 但 UI 返回硬编码数据 |
| **协同过滤** | 🔴 高 | 基于 session 共现关系推荐 |
| **冷启动策略** | 🟡 中 | 无数据时按标签/预定义组合推荐 |
| **推荐反馈** | 🟢 低 | 用户标记推荐有用/无用, 反馈到引擎 |

### 7. 记忆同步

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **Memories 发现** | 🟡 中 | `~/.claude/memories/` 目录已配置, 适配器需支持 |
| **记忆冲突策略** | 🟡 中 | merge/手动选择, 不中心化覆盖 |
| **项目记忆** | 🟢 低 | 项目级 `CLAUDE.md` 随 Git 走, 不归本系统 |

### 8. 服务端增强

| 功能 | 优先级 | 说明 |
|------|:--:|------|
| **案例 API** | 🔴 高 | POST/GET `/api/v1/examples`, `/api/v1/cases` |
| **使用记录 API** | 🔴 高 | POST `/api/v1/usage/batch` (已有端点, 需接入) |
| **组合/链 API** | 🟡 中 | GET/POST `/api/v1/combinations`, `/api/v1/chains` (已有端点, 缺前端) |

---

## 技术债务

| 项目 | 优先级 | 说明 |
|------|:--:|------|
| client 包单测 | 🟡 中 | SyncEngine 缺少单元测试 |
| GitHub Push | 🔴 高 | 解决认证问题, 推送到 fuermos/skill-manage |
| WebView2 兼容 | 🟡 中 | 当前用 native loader, 后续适配新 loader |
| Trae 配置路径确认 | 🟢 低 | 不同系统路径可能不同, 需多平台验证 |

---

## 预计工作量

| 版本 | 内容 | 预估 |
|------|------|------|
| v1.0 | 当前已完成 | - |
| v1.1 | 案例系统 + 自动收集 + 推荐接入 | 3天 |
| v1.2 | 组合/链 UI + 记忆同步 | 2天 |
| v1.3 | 可视化 + 反向增强 + 安装脚本 | 2天 |