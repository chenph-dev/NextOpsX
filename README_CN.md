# NextOpsX（下一代运维平台）

[English](README.md) | **简体中文**

NextOpsX 是面向个人运维 / SRE 的桌面运维工作台，覆盖 SSH、SFTP、数据库、Redis 与 Elasticsearch。它基于 MyGo 原生 UI 构建：单一 Go 代码库，不依赖 WebView。

## 特点

- **一个常驻工作台，而不是分散的功能页** — 资产、终端、数据库控制台与智能运维共享同一个窗口。
- **本地优先** — 资产、命令历史与凭据默认存储在本机；敏感信息使用 AES-256-GCM 加密。
- **执行可控** — 智能运维建议的命令遵循审批策略，模型调用、工具调用与命令执行均有审计记录。
- **原生桌面集成** — 原生菜单、托盘、通知、主题、界面缩放与中英文界面。

## 功能

### 资产与连接

- 资产树支持目录与主机的新增、编辑、移动、重命名和连接测试
- 协议选择面板提供各协议默认值与字段模板
- Windows 资产可拉起系统远程桌面客户端（仅限在 Windows 上运行应用时）
- 重启后恢复上次的布局与选中资产

### 终端与文件

- SSH 多终端页签；切换资产时连接随页签保留
- SFTP 右侧抽屉，支持文件浏览与传输任务
- 命令历史与收藏：记录智能运维执行和应用内触发的命令；终端输入时弹出合并去重后的建议（↑/↓ 选择、Tab 填入），支持搜索、查看输出、重新执行，收藏模板支持 `{{host}}`、`{{user}}` 等变量

### 数据库、缓存与搜索

- 数据库工作台（MySQL、PostgreSQL、SQLite、ClickHouse、SQL Server）：对象树、多标签 SQL、结果网格、分页与 CSV 导出
- Redis：键空间、键详情、TTL、命令终端与 INFO
- Elasticsearch：索引、Query DSL、REST 终端与集群信息

### 智能运维

- 多模型供应商，支持流式回复
- 命令与计划审批；计划批准后可通过普通命令继续执行
- 历史会话与 AI 调用审计
- 运维技能：ZIP 安装、启停、删除与 `SKILL.md` 预览

使用智能运维时，会话上下文会发送给所配置的模型供应商。

### 桌面体验

- 审批策略、桌面通知、原生菜单与托盘
- 系统 / 浅色 / 深色主题、界面缩放与中英文界面
- 从 GitHub Releases 检查并安装应用内更新

## 支持的连接

| 协议 | 能力 |
| --- | --- |
| SSH | 交互终端（多页签）、SFTP、命令执行、智能运维 |
| WinRM | 命令控制台、智能运维 |
| MySQL / PostgreSQL / ClickHouse / SQL Server | SQL 工作台、对象树、CSV 导出 |
| SQLite | 本地数据库工作台 |
| Redis | 键空间、键详情、TTL、命令终端、INFO |
| Elasticsearch | 索引、Query DSL、REST 终端、集群信息 |

## 环境要求

- Go 1.27.3+
- Linux 桌面需要 GTK 运行库；系统托盘还需要 `libayatana-appindicator3`（缺失时仅托盘不可用）
- 打包：Windows 安装包需要 NSIS（在 Windows 上由 `mygo build` 自动获取）；macOS 签名与 DMG 打包需要 macOS

## 快速开始

从源码运行：

```bash
go run .
```

开发时使用热重载：

```bash
go tool mygo dev
```

编译并检查：

```bash
go build ./...
go test ./...
go tool mygo vet .
```

## 打包安装包

```bash
go tool mygo build                    # 当前平台
go tool mygo build -platform linux/amd64,linux/arm64,windows/amd64,windows/arm64,darwin/arm64,darwin/amd64
```

产物写到 `build/`：

- Linux：可执行文件、`.deb`、`.tar.gz`，以及独立的 `install.sh`
- Windows：可执行文件与 NSIS 安装包
- macOS：`.app` 与 `.dmg`（签名与公证需要 macOS）
- 配置 `updates` 后：签名更新包，以及每个构建目标对应的 `update-<platform>.json`

## 数据与凭据

默认使用平台配置目录下的 `nextopsx/` 目录：

| 平台 | 路径 |
| --- | --- |
| Linux | `~/.config/nextopsx` |
| Windows | `%APPDATA%\nextopsx` |
| macOS | `~/Library/Application Support/nextopsx` |

SQLite 数据库为 `nextopsx.db`。

- `NEXTOPSX_DATA_DIR` 可覆盖数据目录，便于测试或便携部署。
- 密码与私钥使用 AES-256-GCM 加密。主密钥优先存入系统 keyring，失败时回退到 `master.key` 文件（权限 `0600`）。

## 自动更新

### 安装更新

通过 `go tool mygo build` 构建且配置了 `updates` 的版本可以自更新：打开 **设置 → 关于**，检查新版本、下载并重启。更新清单来自 GitHub Releases 的最新发布（latest），下载内容会用构建内置的公钥校验，通过后原地替换应用。

应用内更新要求安装目录可写。由系统包管理器安装的版本（例如 `/opt` 下的 `.deb`）由包管理器负责更新，应用会提示不支持自更新。

### 发布版本

签名密钥对由 `mygo keygen` 生成：

```
<配置目录>/mygo/update-keys/mygo-update.key   # 私钥；务必离线备份，绝不提交
<配置目录>/mygo/update-keys/mygo-update.pub   # 公钥；已写入 mygo.json
```

`<配置目录>` 即 MyGo 使用的平台配置目录；Linux 下为 `~/.config`。

在本机发布（需要 `gh` 已登录，或设置 `GH_TOKEN` / `GITHUB_TOKEN`）：

```bash
MYGO_UPDATER_PRIVATE_KEY="$(cat ~/.config/mygo/update-keys/mygo-update.key)" \
  go tool mygo build -platform linux/amd64,linux/arm64,windows/amd64,windows/arm64,darwin/arm64,darwin/amd64 -upload
```

私钥丢失会导致已安装用户无法再收到更新；私钥泄露则任何人都能发布被接受的更新。

CI 发布流程：

1. 把私钥加入仓库 secret：`gh secret set MYGO_UPDATER_PRIVATE_KEY < ~/.config/mygo/update-keys/mygo-update.key`
2. 更新 `mygo.json` 的 `version` 与 `CHANGELOG.md` 的对应小节，提交。
3. 打 tag 并推送：`git tag v0.2.0 && git push origin v0.2.0`
4. `.github/workflows/release.yml` 会校验 tag 与 `mygo.json` 是否一致，在 Ubuntu 上构建 Linux / Windows 目标、在 macOS 上构建 macOS 目标，并把产物上传到 draft release。检查后**发布该 release 并标记为 latest**。
5. `.github/workflows/ci.yml` 在 push 与 PR 时运行 build、test 与 vet。

更新清单与构建目标一一对应（如 `update-linux-amd64.json`），清单地址在构建时写入二进制。发布的架构必须与清单一致：使用 `darwin/universal` 构建时，必须同时提供 `update-darwin-universal.json`，否则该版本无法检查更新。本仓库按 `darwin/arm64` 与 `darwin/amd64` 分别构建和发布。

## 开发

### 可选的 SSH 终端测试

```bash
NEXTOPSX_SSH_TEST_HOST=<host> \
NEXTOPSX_SSH_TEST_USER=<user> \
NEXTOPSX_SSH_TEST_KEY=$HOME/.ssh/<key> \
go test -run TestSSHTerminal -v ./service/terminal
```

### 提交信息钩子

仓库会在提交信息中过滤自动生成的协作者署名与生成声明。克隆仓库后启用一次：

```bash
git config core.hooksPath .githooks
```

钩子位于 `.githooks/commit-msg`，普通提交不受影响；需要保留原始信息时可使用 `git commit --no-verify`。

## 项目结构

```text
cmd/nextopsx       组合根：窗口、菜单、托盘与图标
ui/workspace       UI 状态与动作（ViewModel），不渲染、不直接做 I/O
ui/views           窗口布局（Shell）与功能面板
ui/views/ai        智能运维面板
ui/views/db        数据库工作台
ui/views/redis     Redis 面板
ui/views/es        Elasticsearch 面板
ui/views/sftp      SFTP 抽屉
ui/views/hosts     资产树、表单与欢迎页
ui/views/terminal  终端补全与命令历史面板
ui/views/winrm     WinRM 控制台
ui/views/settings  设置页各分区
ui/views/result    通用结果网格
ui/markdown        原生 Markdown 渲染
ui/ansi            ANSI SGR 解析与终端配色
ui/style           界面缩放
ui/widgets         共享状态文本与终端输出
ui/assets          内嵌图标资源
i18n               zh-CN / en-US 字符串目录
service/backend    功能服务组合根
service/hosts      资产、目录与连接测试
service/ai         AI 会话与供应商
service/db         数据库查询与结构
service/sftp       SFTP 浏览与传输任务
service/skills     运维技能
service/snippets   命令收藏
service/settings   UI 设置
service/audit      AI 调用审计
service/policy     审批策略
service/history    命令历史
service/rdp        外部 RDP 拉起
service/terminal   交互终端适配
service/executor   连接执行器解析
domain             连接器、执行器、SFTP 与技能
agent              AI 会话、模型、工具与提示词
store              SQLite、加密、迁移与领域存储
resources          应用图标源文件
```
