// Package backend wires the framework-independent NextOpsX services.
// Nothing in this package imports a GUI framework; the native UI calls it
// directly.
package backend

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"

	"nextopsx/internal/einoagent/session"
	"nextopsx/internal/handler/aiconfig"
	"nextopsx/internal/handler/approvalpolicy"
	"nextopsx/internal/handler/base"
	connectorhandler "nextopsx/internal/handler/connector"
	dbhandler "nextopsx/internal/handler/db"
	logshandler "nextopsx/internal/handler/logs"
	rdphandler "nextopsx/internal/handler/rdp"
	sessionshandler "nextopsx/internal/handler/sessions"
	sftppkg "nextopsx/internal/sftp"
	"nextopsx/internal/skill"
	"nextopsx/internal/sshexec"
	"nextopsx/internal/store"
	configstore "nextopsx/internal/store/config"
	convstore "nextopsx/internal/store/conversations"
	hoststore "nextopsx/internal/store/hosts"
	logsstore "nextopsx/internal/store/logs"
	skillsstore "nextopsx/internal/store/skills"
)

// Backend owns the stores, executors and AI services of the app.
type Backend struct {
	db       *store.DB
	hosts    *hoststore.HostsStore
	cfg      *configstore.ConfigStore
	policy   *configstore.PolicyStore
	convs    *convstore.ConvStore
	logs     *logsstore.LogsStore
	skills   *skill.Manager
	resolver *base.ExecutorResolver
	sm       *session.SessionManager
	sftp     *sftppkg.Manager

	aiConfig   *aiconfig.AIConfigHandler
	approval   *approvalpolicy.ApprovalPolicyHandler
	dbHandler  *dbhandler.DbHandler
	connectors *connectorhandler.ConnectorHandler
	rdp        *rdphandler.RdpHandler
	logsAPI    *logshandler.LogsHandler
	sessions   *sessionshandler.SessionsHandler

	// aiSink 接收 SessionManager 的 AI 事件（前台 UI 注入）。
	aiMu   sync.Mutex
	aiSink AIEventSink

	// sftpSink 接收传输任务事件（前台 UI 注入）。
	sftpMu   sync.Mutex
	sftpSink SFTPEventSink
}

// SFTPEventSink receives transfer task events: "sftp:start" when a task is
// created, "sftp:progress" on progress and final state.
type SFTPEventSink func(event string, task sftppkg.TaskInfo)

// SetSFTPEventSink installs the transfer event sink (nil removes it).
func (b *Backend) SetSFTPEventSink(sink SFTPEventSink) {
	b.sftpMu.Lock()
	b.sftpSink = sink
	b.sftpMu.Unlock()
}

func (b *Backend) emitSFTP(event string, t *sftppkg.Task) {
	b.sftpMu.Lock()
	sink := b.sftpSink
	b.sftpMu.Unlock()
	if sink != nil {
		sink(event, t.Snapshot())
	}
}

// AIEventSink 接收一条 AI 会话事件。sessionID 标识会话，event 为
// "ai:text"/"ai:command"/"ai:plan"/"run:*"/"session:state"/"ai:error"，
// data 为对应载荷（结构体或 map）。实现方通常把事件切到 UI 线程处理。
type AIEventSink func(sessionID, event string, data any)

// SetAIEventSink 注入 AI 事件接收器（可传 nil 取消）。
func (b *Backend) SetAIEventSink(sink AIEventSink) {
	b.aiMu.Lock()
	b.aiSink = sink
	b.aiMu.Unlock()
}

// New opens the data store and wires every service.
func New() (*Backend, error) {
	db, err := store.Open()
	if err != nil {
		return nil, err
	}
	cfg := configstore.NewConfigStore(db)
	policy := configstore.NewPolicyStore(db)
	hosts := hoststore.NewHostsStore(db)
	convs := convstore.NewConvStore(db)
	logs := logsstore.NewLogsStore(db)

	dataDir, err := store.DataDir()
	if err != nil {
		return nil, err
	}
	skillRoot := filepath.Join(dataDir, "skills")
	if err := os.MkdirAll(skillRoot, 0o700); err != nil {
		return nil, err
	}
	skillMgr := skill.NewManager(skillsstore.NewSkillStore(db), skillRoot)
	resolver := base.NewExecutorResolver(hosts)

	b := &Backend{
		db: db, hosts: hosts, cfg: cfg, policy: policy, convs: convs,
		logs: logs, skills: skillMgr, resolver: resolver,
	}
	b.sm = session.NewSessionManager(db, cfg, resolver.ExecFor,
		func(hostID string) (string, error) {
			meta, err := hosts.HostMetaByID(hostID)
			if err != nil {
				return "", err
			}
			return meta.Name, nil
		},
		b.emitAI,
	)
	b.sm.SetApprovalPolicyResolver(func(hostID string) (bool, []string) {
		p, err := policy.GetApprovalPolicy()
		if err != nil {
			return false, nil
		}
		auto := p.EnableAuto
		if override, oerr := hosts.GetAutoApprove(hostID); oerr == nil {
			switch override {
			case "on":
				auto = true
			case "off":
				auto = false
			}
		}
		return auto, p.ReadOnlyList
	})
	b.sm.SetSkillResolver(skillMgr.Catalog, skillMgr.Lookup)
	b.sm.SetProtocolResolver(func(hostID string) string {
		meta, err := hosts.HostMetaByID(hostID)
		if err != nil || meta == nil {
			return "ssh"
		}
		return meta.Protocol
	})
	b.sm.SetCapabilityResolver(resolver.ConnFor)

	b.sftp = sftppkg.NewManager(
		resolver.HostFor,
		func(t *sftppkg.Task) { b.emitSFTP("sftp:progress", t) },
		func(t *sftppkg.Task) { b.emitSFTP("sftp:start", t) },
	)
	b.aiConfig = aiconfig.NewAIConfigHandler(cfg, b.sm.InvalidateConfig)
	b.approval = approvalpolicy.NewApprovalPolicyHandler(policy, b.sm.InvalidateConfig)
	b.dbHandler = dbhandler.NewDbHandler(resolver)
	b.connectors = connectorhandler.NewConnectorHandler()
	b.rdp = rdphandler.NewRdpHandler(hosts)
	b.logsAPI = logshandler.NewLogsHandler(logs)
	b.sessions = sessionshandler.NewSessionsHandler(convs, b.sm)
	return b, nil
}

// ListHosts returns the asset tree.
func (b *Backend) ListHosts() ([]hoststore.TreeNode, error) {
	return b.hosts.ListTree()
}

// OpenTerminalConn opens an interactive SSH session for hostID and adapts
// it to the terminal plugin's io.ReadWriteCloser. Only SSH assets support
// an interactive terminal.
func (b *Backend) OpenTerminalConn(hostID string, cols, rows int) (io.ReadWriteCloser, error) {
	meta, err := b.hosts.HostMetaByID(hostID)
	if err != nil {
		return nil, err
	}
	if meta == nil {
		return nil, fmt.Errorf("资产不存在")
	}
	proto := meta.Protocol
	if proto != "" && proto != "ssh" {
		return nil, fmt.Errorf("仅 SSH 资产支持交互式终端（当前：%s）", proto)
	}
	secret, authType, err := b.hosts.GetHostSecret(hostID)
	if err != nil {
		return nil, fmt.Errorf("读取凭据失败: %w", err)
	}
	host := sshexec.Host{
		Addr: meta.Addr, Port: meta.Port, User: meta.User,
		AuthType: authType, Secret: secret,
	}
	// TOFU: pin the host key on first connect, refuse a changed key.
	host.TrustHostKey = func(fp string) error {
		cur, err := b.hosts.HostKeyFingerprint(hostID)
		if err != nil {
			return err
		}
		if cur == "" {
			return b.hosts.SaveHostKeyFingerprint(hostID, fp)
		}
		if cur != fp {
			return fmt.Errorf("主机密钥已变更（%s → %s），可能遭中间人替换", cur, fp)
		}
		return nil
	}

	sess, err := sshexec.OpenSession(context.Background(), host, cols, rows)
	if err != nil {
		return nil, err
	}
	return newSSHConn(sess), nil
}

// Close releases the SFTP clients and the SQLite connection.
func (b *Backend) Close() {
	if b == nil {
		return
	}
	if b.sftp != nil {
		b.sftp.Close()
	}
	if b.db != nil {
		if sqlDB, err := b.db.GORM().DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}

// emitAI is the SessionManager's event sink. It forwards to the native AI
// panel when one is attached, and drops events otherwise.
func (b *Backend) emitAI(sessionID, event string, data any) {
	b.aiMu.Lock()
	sink := b.aiSink
	b.aiMu.Unlock()
	if sink != nil {
		sink(sessionID, event, data)
	}
}
