// Package backend wires the framework-independent NextOpsX services.
// Nothing in this package imports a GUI framework; the native UI calls the
// feature services it exposes.
package backend

import (
	"os"
	"path/filepath"
	"sync"

	"nextopsx/agent/session"
	sftppkg "nextopsx/domain/sftp"
	"nextopsx/domain/skill"
	"nextopsx/service/ai"
	"nextopsx/service/audit"
	dbsvc "nextopsx/service/db"
	"nextopsx/service/executor"
	"nextopsx/service/history"
	"nextopsx/service/hosts"
	"nextopsx/service/policy"
	"nextopsx/service/rdp"
	"nextopsx/service/settings"
	"nextopsx/service/sftp"
	"nextopsx/service/skills"
	"nextopsx/service/snippets"
	"nextopsx/service/terminal"
	"nextopsx/store"
	configstore "nextopsx/store/config"
	convstore "nextopsx/store/conversations"
	historystore "nextopsx/store/history"
	hoststore "nextopsx/store/hosts"
	logsstore "nextopsx/store/logs"
	skillsstore "nextopsx/store/skills"
	snippetsstore "nextopsx/store/snippets"
)

// Backend owns the stores and the feature services built over them.
type Backend struct {
	Hosts    *hosts.Service
	AI       *ai.Service
	DB       *dbsvc.Service
	SFTP     *sftp.Service
	Skills   *skills.Service
	Snippets *snippets.Service
	Settings *settings.Service
	Audit    *audit.Service
	Policy   *policy.Service
	History  *history.Service
	RDP      *rdp.Service
	Terminal *terminal.Service
	Executor *executor.Resolver

	db *store.DB

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
	data, err := store.Open()
	if err != nil {
		return nil, err
	}
	providers := configstore.NewProviderStore(data)
	policyStore := configstore.NewPolicyStore(data)
	uiSettings := configstore.NewUISettingsStore(data)
	hostsStore := hoststore.NewHostsStore(data)
	convs := convstore.NewConvStore(data)
	logs := logsstore.NewLogsStore(data)
	historyStore := historystore.NewStore(data)
	snippetStore := snippetsstore.NewSnippetStore(data)

	dataDir, err := store.DataDir()
	if err != nil {
		return nil, err
	}
	skillRoot := filepath.Join(dataDir, "skills")
	if err := os.MkdirAll(skillRoot, 0o700); err != nil {
		return nil, err
	}
	skillMgr := skill.NewManager(skillsstore.NewSkillStore(data), skillRoot)
	resolver := executor.NewResolver(hostsStore)

	b := &Backend{db: data, Executor: resolver}

	sm := session.NewSessionManager(data, providers, resolver.ExecFor,
		func(hostID string) (string, error) {
			meta, err := hostsStore.HostMetaByID(hostID)
			if err != nil {
				return "", err
			}
			return meta.Name, nil
		},
		b.emitAI,
	)
	sm.SetApprovalPolicyResolver(func(hostID string) (bool, []string) {
		p, err := policyStore.GetApprovalPolicy()
		if err != nil {
			return false, nil
		}
		auto := p.EnableAuto
		if override, oerr := hostsStore.GetAutoApprove(hostID); oerr == nil {
			switch override {
			case "on":
				auto = true
			case "off":
				auto = false
			}
		}
		return auto, p.ReadOnlyList
	})
	sm.SetSkillResolver(skillMgr.Catalog, skillMgr.LookupEnabled)
	sm.SetProtocolResolver(func(hostID string) string {
		meta, err := hostsStore.HostMetaByID(hostID)
		if err != nil || meta == nil {
			return "ssh"
		}
		return meta.Protocol
	})
	sm.SetCapabilityResolver(resolver.ConnFor)

	b.Hosts = hosts.New(hostsStore, resolver, sm.InvalidateConfig)
	b.AI = ai.New(sm, convs, providers, sm.InvalidateConfig)
	b.DB = dbsvc.New(resolver)
	b.SFTP = sftp.New(sftppkg.NewManager(
		resolver.HostFor,
		func(t *sftppkg.Task) { b.emitSFTP("sftp:progress", t) },
		func(t *sftppkg.Task) { b.emitSFTP("sftp:start", t) },
	))
	b.Skills = skills.New(skillMgr, sm.InvalidateConfig)
	b.Snippets = snippets.New(snippetStore)
	b.Settings = settings.New(uiSettings)
	b.Audit = audit.New(logs)
	b.Policy = policy.New(policyStore, sm.InvalidateConfig)
	b.History = history.New(historyStore)
	b.RDP = rdp.New(hostsStore)
	b.Terminal = terminal.New(hostsStore)
	return b, nil
}

// Close releases the SFTP clients, the executor cache and the SQLite connection.
func (b *Backend) Close() {
	if b == nil {
		return
	}
	if b.SFTP != nil {
		b.SFTP.Close()
	}
	if b.Executor != nil {
		b.Executor.Close()
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
