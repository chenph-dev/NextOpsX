# NextOpsX（下一代运维平台）

NextOpsX 是基于 MyGo 原生 UI 的桌面运维客户端：不依赖 WebView，保持单一 Go 代码库。

## 功能

- 资产树与资产管理：目录/主机 CRUD、移动、重命名、连接测试
- SSH 交互终端（libghostty-vt）、SFTP 右侧抽屉与传输任务
- 智能运维全协议底部抽屉；布局与上次选中资产在重启后恢复
- 数据库工作台：对象树、多标签 SQL、结果网格、分页、CSV 导出（MySQL / PostgreSQL / SQLite / ClickHouse / SQL Server）
- Redis：键空间、键详情、TTL、命令终端、INFO
- Elasticsearch：索引、Query DSL、REST 终端、集群信息
- 智能运维：多供应商、流式回复、命令/计划审批、历史会话、调用审计
- 运维技能：ZIP 安装、启停、删除、SKILL.md 预览
- 命令历史与收藏：记录智能运维执行和应用内触发的命令，终端输入时弹出合并去重后的历史/收藏建议（↑↓ 选择、Tab 填入）；支持搜索、查看输出、重新执行与收藏，收藏模板支持 `{{host}}`、`{{user}}` 等变量
- 审批策略、桌面通知、原生菜单、托盘、主题与界面缩放、中英文界面

## 环境要求

- Go 1.27.3+
- Linux 桌面需要 GTK 运行库；系统托盘需要 `libayatana-appindicator3`（缺失时仅托盘不可用）
- 打包 Windows 安装包需要 NSIS（在 Windows 上由 `mygo build` 自动获取）；macOS 签名与 dmg 需要 macOS

## 开发

```bash
go run .                 # 启动应用
go tool mygo dev         # 热重载开发
go build ./...           # 编译所有包
go test ./...            # 运行测试
go tool mygo vet .       # MyGo UI 生命周期检查
```

真实 SSH 终端测试是可选的：

```bash
NEXTOPSX_SSH_TEST_HOST=<host> \
NEXTOPSX_SSH_TEST_USER=<user> \
NEXTOPSX_SSH_TEST_KEY=$HOME/.ssh/<key> \
go test -run TestSSHTerminal -v ./service/backend
```

## 数据目录

默认使用平台配置目录下的 `nextopsx/`（Linux：`~/.config/nextopsx`，Windows：`%APPDATA%/nextopsx`），数据库为 `nextopsx.db`。

- `NEXTOPSX_DATA_DIR` 可覆盖数据目录，便于测试或便携部署。
- 密码/私钥使用 AES-256-GCM 加密，优先存入系统 keyring，失败时回退到 `master.key`。

## 打包

```bash
go tool mygo build                                   # 当前平台
go tool mygo build -platform linux/amd64,windows/amd64,darwin/universal
```

产物写到 `build/`：

- Linux：可执行文件、`.deb`、`.tar.gz`（含 `install.sh`）
- Windows：可执行文件、NSIS 安装包
- macOS：`.app` 与 `.dmg`（需在 macOS 上签名/公证）
- 配置 `updates` 后还会生成签名更新包与 `update-<platform>.json`

## 自动更新

`mygo.json` 已配置 GitHub Releases 更新源（`chenph-dev/NextOpsX`）和公钥；安装版可在「设置 → 关于」检查、下载并重启应用更新。

签名私钥由 `mygo keygen` 生成在本机：

```
~/.config/mygo/update-keys/mygo-update.key   # 私钥，务必离线备份，绝不提交
~/.config/mygo/update-keys/mygo-update.pub   # 公钥，已写入 mygo.json
```

发布新版本（需要 `gh` 已登录，或设置 `GH_TOKEN`/`GITHUB_TOKEN`）：

```bash
MYGO_UPDATER_PRIVATE_KEY="$(cat ~/.config/mygo/update-keys/mygo-update.key)" \
  go tool mygo build -platform linux/amd64,windows/amd64,darwin/arm64 -upload
```

私钥丢失会导致已安装用户无法再收到更新；泄露则任何人都能发布被接受的更新。

CI 发布流程：

1. 把私钥内容加入仓库 secret：`gh secret set MYGO_UPDATER_PRIVATE_KEY < ~/.config/mygo/update-keys/mygo-update.key`
2. 更新 `mygo.json` 的 `version` 和 `CHANGELOG.md` 对应小节，提交。
3. 打 tag 并推送：`git tag v0.2.0 && git push origin v0.2.0`
4. `release.yml` 会在 Linux / Windows / macOS runner 上构建并上传到 draft release；检查后发布。
5. `ci.yml` 在 push/PR 时运行 build、test 与 vet。

## 项目结构

```text
cmd/nextopsx      # 组合根：窗口、菜单、托盘、tray 图标
ui/app            # UI 状态与动作
ui/views          # MyGo 原生视图
ui/assets         # 内嵌图标资源
i18n              # zh-CN / en-US 字符串目录
service/backend   # 服务装配（不依赖 MyGo）
domain            # 连接器、执行器、SFTP、技能
agent             # AI 会话、模型、工具与提示词
store             # SQLite、加密、迁移与各领域存储
```
