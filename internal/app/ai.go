package app

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"nextopsx/internal/einoagent/session"
	configstore "nextopsx/internal/store/config"
	convstore "nextopsx/internal/store/conversations"
)

// AI 审批状态（与后端落库的 approval_status 一致）。
const (
	AIStatusPending  = "pending"
	AIStatusApproved = "approved"
	AIStatusRejected = "rejected"
	AIStatusAuto     = "auto"
)

// CommandSuggestion 是模型提议执行的一条命令（审批卡数据）。
type CommandSuggestion struct {
	Command      string `json:"command"`
	Why          string `json:"why"`
	Risk         string `json:"risk"`
	AssessedRisk string `json:"assessedRisk"`
}

// PlanSuggestion 是模型提交的执行计划（create_plan）。
type PlanSuggestion struct {
	Goal  string   `json:"goal"`
	Steps []string `json:"steps"`
}

// aiTurnSnapshot keeps a non-active session's live UI state, so switching
// conversations or assets does not lose a pending approval or error.
type aiTurnSnapshot struct {
	Streaming      string
	PendingCommand *CommandSuggestion
	CommandStatus  string
	CommandDraft   string
	CommandEditing bool
	CommandConfirm bool
	PendingPlan    *PlanSuggestion
	PlanStatus     string
	State          string
	Err            string
	ErrCancelled   bool
	RunningCommand string
	RunStartedAt   time.Time
	RunOutput      string
}

// AIState 是 AI 对话面板的状态。视图只读取它，动作方法写它。
type AIState struct {
	HostID    string
	SessionID string

	Conversations []convstore.Conversation
	Messages      []convstore.Message
	Streaming     string

	PendingCommand *CommandSuggestion
	CommandStatus  string
	CommandDraft   string
	CommandEditing bool
	CommandConfirm bool
	PendingPlan    *PlanSuggestion
	PlanStatus     string

	State        string
	Err          string
	ErrCancelled bool

	RunningCommand string
	RunStartedAt   time.Time
	RunOutput      string

	Input      string
	Configured bool
	Loading    bool

	// 当前对话使用的供应商（运行时；编辑草稿在 SettingsState）。
	Providers      []configstore.AIProvider
	ProviderID     string
	ProviderName   string
	ProviderChoice string // 绑定给供应商下拉框，切换后回写后端
	Model          string

	// 历史会话弹层
	HistoryOpen   bool
	RenameID      string
	RenameTitle   string
	DeleteConfirm string

	// 用于折叠工具输出块：消息 ID → 是否展开。
	ToolOpen map[string]bool

	// Scroll 保持消息区的滚动位置；靠近底部时自动跟随新消息。
	Scroll ui.ScrollState

	attachSeq    uint64
	resyncSeq    uint64
	commandEpoch uint64
	planEpoch    uint64
	providerSeq  uint64 // 使过期的供应商读取失效
	switching    bool   // 供应商切换请求在途（串行写入，保证最后一次选择生效）
	switchTarget string // 最新的切换目标
	localSeq     int
	approveBusy  bool

	snapshots map[string]*aiTurnSnapshot
}

// Busy reports whether a turn is generating or running a command.
func (s *AIState) Busy() bool {
	return s.State == session.StateThinking || s.State == session.StateRunning
}

// InputDisabled reports whether the composer must be disabled.
func (s *AIState) InputDisabled() bool {
	return !s.Configured || s.Busy() || s.State == session.StateAwaitingApproval
}

// SelectedProtocol returns the selected asset's protocol, defaulting to ssh.
func (a *App) SelectedProtocol() string {
	if a.Selected == nil || a.Selected.Protocol == "" {
		return "ssh"
	}
	return a.Selected.Protocol
}

// attachAI loads (and lazily creates) the AI conversation for a host.
func (a *App) attachAI(hostID string) {
	a.saveAISnapshot()
	a.AI.attachSeq++
	seq := a.AI.attachSeq
	host := hostID
	a.resetAIForHost(host)
	if a.Backend == nil {
		a.AI.Loading = false
		return
	}
	b := a.Backend
	go func() {
		sid, err := b.EnsureAISession(host)
		var msgs []convstore.Message
		var prov configstore.AIProvider
		provOK := false
		if err == nil {
			msgs, _ = b.LoadAIMessages(sid)
			if p, perr := b.ProviderForConversation(sid); perr == nil {
				prov, provOK = p, true
			}
		}
		convs, convErr := b.ListAIConversations(host)
		providers, providersErr := b.ListAIProviders()
		a.Post(func() {
			if seq != a.AI.attachSeq {
				return
			}
			a.AI.Loading = false
			if convErr == nil {
				a.AI.Conversations = convs
			}
			if providersErr == nil {
				a.AI.Providers = providers
			}
			a.AI.providerSeq++
			a.applyAIProvider(prov, provOK)
			if err != nil {
				a.AI.Err = "打开 AI 会话失败：" + err.Error()
				return
			}
			a.AI.SessionID = sid
			a.restoreAISnapshot(sid)
			a.AI.Messages = msgs
		})
	}()
}

func (a *App) resetAIForHost(hostID string) {
	a.AI.HostID = hostID
	a.AI.SessionID = ""
	a.AI.Conversations = nil
	a.AI.Messages = nil
	a.AI.Streaming = ""
	a.AI.PendingCommand = nil
	a.AI.CommandStatus = ""
	a.AI.CommandDraft = ""
	a.AI.CommandEditing = false
	a.AI.CommandConfirm = false
	a.AI.PendingPlan = nil
	a.AI.PlanStatus = ""
	a.AI.State = ""
	a.AI.Err = ""
	a.AI.ErrCancelled = false
	a.AI.RunningCommand = ""
	a.AI.RunStartedAt = time.Time{}
	a.AI.RunOutput = ""
	a.AI.Input = ""
	a.AI.Loading = true
	a.AI.HistoryOpen = false
	a.AI.RenameID = ""
	a.AI.RenameTitle = ""
	a.AI.DeleteConfirm = ""
	a.AI.ToolOpen = map[string]bool{}
	a.AI.approveBusy = false
	a.AI.Scroll.Y = 0
	a.AI.Scroll.MaxY = 0
	// 供应商信息随会话异步加载；先清空，避免短暂显示上一个资产的供应商。
	a.AI.providerSeq++
	a.AI.Providers = nil
	a.AI.ProviderID = ""
	a.AI.ProviderName = ""
	a.AI.ProviderChoice = ""
	a.AI.switchTarget = ""
	a.AI.switching = false
	a.AI.Model = ""
	a.AI.Configured = false
}

// HandleAIEvent is the backend's AI event sink. Events arrive on worker
// goroutines, so they are republished to the UI thread.
func (a *App) HandleAIEvent(sessionID, event string, data any) {
	a.Post(func() {
		a.applyAIEvent(sessionID, event, data)
	})
}

func (a *App) applyAIEvent(sessionID, event string, data any) {
	if sessionID == "" {
		return
	}
	if sessionID != a.AI.SessionID {
		a.applyAIEventSnapshot(sessionID, event, data)
		return
	}
	switch event {
	case "ai:text":
		var p struct {
			Delta string `json:"delta"`
		}
		if decodeAIEvent(data, &p) && p.Delta != "" {
			a.AI.Streaming += p.Delta
		}
	case "ai:command":
		var p CommandSuggestion
		if decodeAIEvent(data, &p) {
			a.AI.commandEpoch++
			a.AI.PendingCommand = &p
			a.AI.CommandStatus = AIStatusPending
			a.AI.CommandDraft = p.Command
			a.AI.CommandEditing = false
			a.AI.CommandConfirm = false
			a.AI.Streaming = ""
		}
	case "ai:plan":
		var p PlanSuggestion
		if decodeAIEvent(data, &p) {
			a.AI.planEpoch++
			a.AI.PendingPlan = &p
			a.AI.PlanStatus = AIStatusPending
			a.AI.Streaming = ""
		}
	case "run:auto", "run:start":
		var p struct {
			Command string `json:"command"`
		}
		if decodeAIEvent(data, &p) {
			a.AI.RunningCommand = p.Command
			a.AI.RunStartedAt = time.Now()
			a.AI.RunOutput = ""
		}
	case "run:output":
		var p struct {
			Delta string `json:"delta"`
		}
		if decodeAIEvent(data, &p) && p.Delta != "" {
			a.AI.RunOutput += p.Delta
		}
	case "run:result":
		a.AI.RunningCommand = ""
		a.AI.RunStartedAt = time.Time{}
		a.AI.RunOutput = ""
		a.resyncAI(sessionID)
	case "ai:error":
		var p struct {
			Message   string `json:"message"`
			Cancelled bool   `json:"cancelled"`
		}
		if decodeAIEvent(data, &p) {
			a.AI.Err = p.Message
			a.AI.ErrCancelled = p.Cancelled
			a.AI.Streaming = ""
		}
	case "session:state":
		if st, ok := data.(string); ok {
			prev := a.AI.State
			a.AI.State = st
			if st == session.StateIdle {
				if prev == session.StateThinking || prev == session.StateRunning {
					a.notifyAITurnEnded()
				}
				a.resyncAI(sessionID)
			}
		}
	}
	a.saveAISnapshot()
}

// notifyAITurnEnded reports a finished turn. Cancellations are the user's own
// doing and stay silent; failures say so instead of claiming a reply.
func (a *App) notifyAITurnEnded() {
	a.notifyTurnResult(a.AI.Err, a.AI.ErrCancelled, a.AI.HostID)
}

// notifyTurnResult reports a finished turn. Cancellations are the user's own
// doing and stay silent; failures say so instead of claiming a reply.
func (a *App) notifyTurnResult(errMsg string, cancelled bool, hostID string) {
	if cancelled {
		return
	}
	host := ""
	if n := findTreeNode(a.Roots, hostID); n != nil {
		host = n.Name
	}
	switch {
	case errMsg != "":
		body := errMsg
		if host != "" {
			body = "「" + host + "」：" + body
		}
		a.notifyUser("AI 对话出错", body)
	case host != "":
		a.notifyUser("AI 对话完成", "「"+host+"」的回复已生成")
	default:
		a.notifyUser("AI 对话完成", "后台对话的回复已生成")
	}
}

// saveAISnapshot copies the active session's live state for a later switch.
func (a *App) saveAISnapshot() {
	sid := a.AI.SessionID
	if sid == "" {
		return
	}
	if a.AI.snapshots == nil {
		a.AI.snapshots = map[string]*aiTurnSnapshot{}
	}
	s := &aiTurnSnapshot{
		Streaming:      a.AI.Streaming,
		CommandStatus:  a.AI.CommandStatus,
		CommandDraft:   a.AI.CommandDraft,
		CommandEditing: a.AI.CommandEditing,
		CommandConfirm: a.AI.CommandConfirm,
		PlanStatus:     a.AI.PlanStatus,
		State:          a.AI.State,
		Err:            a.AI.Err,
		ErrCancelled:   a.AI.ErrCancelled,
		RunningCommand: a.AI.RunningCommand,
		RunStartedAt:   a.AI.RunStartedAt,
		RunOutput:      a.AI.RunOutput,
	}
	if a.AI.PendingCommand != nil {
		c := *a.AI.PendingCommand
		s.PendingCommand = &c
	}
	if a.AI.PendingPlan != nil {
		p := *a.AI.PendingPlan
		p.Steps = append([]string(nil), p.Steps...)
		s.PendingPlan = &p
	}
	a.AI.snapshots[sid] = s
}

// restoreAISnapshot loads a session's live state into the active fields.
func (a *App) restoreAISnapshot(sid string) {
	s := a.AI.snapshots[sid]
	if s == nil {
		a.AI.Streaming = ""
		a.AI.PendingCommand = nil
		a.AI.CommandStatus = ""
		a.AI.CommandDraft = ""
		a.AI.CommandEditing = false
		a.AI.CommandConfirm = false
		a.AI.PendingPlan = nil
		a.AI.PlanStatus = ""
		a.AI.State = ""
		a.AI.Err = ""
		a.AI.ErrCancelled = false
		a.AI.RunningCommand = ""
		a.AI.RunStartedAt = time.Time{}
		a.AI.RunOutput = ""
		return
	}
	a.AI.Streaming = s.Streaming
	a.AI.CommandStatus = s.CommandStatus
	a.AI.CommandDraft = s.CommandDraft
	a.AI.CommandEditing = s.CommandEditing
	a.AI.CommandConfirm = s.CommandConfirm
	a.AI.PlanStatus = s.PlanStatus
	a.AI.State = s.State
	a.AI.Err = s.Err
	a.AI.ErrCancelled = s.ErrCancelled
	a.AI.RunningCommand = s.RunningCommand
	a.AI.RunStartedAt = s.RunStartedAt
	a.AI.RunOutput = s.RunOutput
	a.AI.PendingCommand = nil
	if s.PendingCommand != nil {
		c := *s.PendingCommand
		a.AI.PendingCommand = &c
	}
	a.AI.PendingPlan = nil
	if s.PendingPlan != nil {
		p := *s.PendingPlan
		p.Steps = append([]string(nil), p.Steps...)
		a.AI.PendingPlan = &p
	}
}

// applyAIEventSnapshot records events of a session that is not on screen.
func (a *App) applyAIEventSnapshot(sessionID, event string, data any) {
	if a.AI.snapshots == nil {
		a.AI.snapshots = map[string]*aiTurnSnapshot{}
	}
	s := a.AI.snapshots[sessionID]
	if s == nil {
		s = &aiTurnSnapshot{}
		a.AI.snapshots[sessionID] = s
	}
	switch event {
	case "ai:text":
		var p struct {
			Delta string `json:"delta"`
		}
		if decodeAIEvent(data, &p) && p.Delta != "" {
			s.Streaming += p.Delta
		}
	case "ai:command":
		var p CommandSuggestion
		if decodeAIEvent(data, &p) {
			s.PendingCommand = &p
			s.CommandStatus = AIStatusPending
			s.CommandDraft = p.Command
			s.CommandEditing = false
			s.CommandConfirm = false
			s.Streaming = ""
		}
	case "ai:plan":
		var p PlanSuggestion
		if decodeAIEvent(data, &p) {
			s.PendingPlan = &p
			s.PlanStatus = AIStatusPending
			s.Streaming = ""
		}
	case "run:auto", "run:start":
		var p struct {
			Command string `json:"command"`
		}
		if decodeAIEvent(data, &p) {
			s.RunningCommand = p.Command
			s.RunStartedAt = time.Now()
			s.RunOutput = ""
		}
	case "run:output":
		var p struct {
			Delta string `json:"delta"`
		}
		if decodeAIEvent(data, &p) && p.Delta != "" {
			s.RunOutput += p.Delta
		}
	case "run:result":
		s.RunningCommand = ""
		s.RunStartedAt = time.Time{}
		s.RunOutput = ""
	case "ai:error":
		var p struct {
			Message   string `json:"message"`
			Cancelled bool   `json:"cancelled"`
		}
		if decodeAIEvent(data, &p) {
			s.Err = p.Message
			s.ErrCancelled = p.Cancelled
			s.Streaming = ""
		}
	case "session:state":
		if st, ok := data.(string); ok {
			prev := s.State
			s.State = st
			if st == session.StateIdle &&
				(prev == session.StateThinking || prev == session.StateRunning) {
				// 后台会话完成也要通知；这里拿不到资产名，用通用文案。
				a.notifyTurnResult(s.Err, s.ErrCancelled, "")
			}
		}
	}
}

func decodeAIEvent(data any, out any) bool {
	b, err := json.Marshal(data)
	if err != nil {
		return false
	}
	return json.Unmarshal(b, out) == nil
}

// resyncAI reloads persisted messages and conversation titles. The DB is
// the source of truth after each completed step; live streaming and
// approval cards are cleared only when no newer command proposal arrived
// while loading.
func (a *App) resyncAI(sessionID string) {
	a.AI.resyncSeq++
	seq := a.AI.resyncSeq
	epoch := a.AI.commandEpoch
	planEpoch := a.AI.planEpoch
	host := a.AI.HostID
	if a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		msgs, err := b.LoadAIMessages(sessionID)
		if err != nil {
			return
		}
		convs, _ := b.ListAIConversations(host)
		a.Post(func() {
			if seq != a.AI.resyncSeq || sessionID != a.AI.SessionID {
				return
			}
			a.AI.Messages = msgs
			a.AI.Streaming = ""
			a.AI.RunningCommand = ""
			a.AI.RunStartedAt = time.Time{}
			a.AI.RunOutput = ""
			// Keep approval cards while the backend still waits for a
			// decision; they are the only interactive UI for that state.
			if epoch == a.AI.commandEpoch && a.AI.State != session.StateAwaitingApproval {
				a.AI.PendingCommand = nil
				a.AI.CommandStatus = ""
				a.AI.CommandDraft = ""
				a.AI.CommandEditing = false
				a.AI.CommandConfirm = false
			}
			if planEpoch == a.AI.planEpoch && a.AI.State != session.StateAwaitingApproval {
				a.AI.PendingPlan = nil
				a.AI.PlanStatus = ""
			}
			if convs != nil {
				a.AI.Conversations = convs
			}
			a.saveAISnapshot()
		})
	}()
}

// SendAIMessage sends the composer text (or a slash command).
func (a *App) SendAIMessage() {
	text := strings.TrimSpace(a.AI.Input)
	if text == "" {
		return
	}
	switch text {
	case "/clear":
		a.AI.Input = ""
		a.ClearAIConversation()
		return
	case "/new":
		a.AI.Input = ""
		a.NewAIConversation()
		return
	}
	if a.AI.SessionID == "" || a.AI.InputDisabled() {
		return
	}
	a.AI.Input = ""
	a.AI.Streaming = ""
	a.AI.Err = ""
	a.AI.ErrCancelled = false

	sid := a.AI.SessionID
	a.AI.localSeq++
	localID := fmt.Sprintf("local-%d-%d", time.Now().UnixNano(), a.AI.localSeq)
	a.AI.Messages = append(a.AI.Messages, convstore.Message{
		ID: localID, SessionID: sid, Role: "user", Content: text,
		Ts: time.Now().Unix(),
	})
	// Optimistically show the turn as running so a second send cannot slip
	// in before the backend's first state event.
	a.AI.State = session.StateThinking
	a.saveAISnapshot()
	if a.Backend == nil {
		a.AI.State = session.StateIdle
		return
	}
	b := a.Backend
	go func() {
		if err := b.SendAIMessage(sid, text); err != nil {
			a.Post(func() {
				if sid != a.AI.SessionID {
					return
				}
				a.AI.Messages = removeAIMessage(a.AI.Messages, localID)
				a.AI.Err = "发送失败：" + err.Error()
				a.AI.ErrCancelled = false
				// Some pre-flight failures return before emitting a state
				// event; without this the composer would stay disabled.
				a.AI.State = session.StateIdle
				a.saveAISnapshot()
			})
		}
	}()
}

func removeAIMessage(msgs []convstore.Message, id string) []convstore.Message {
	out := msgs[:0]
	for _, m := range msgs {
		if m.ID != id {
			out = append(out, m)
		}
	}
	return out
}

// ApproveAICommand approves the pending command with the (possibly edited) draft.
func (a *App) ApproveAICommand() {
	if a.AI.PendingCommand == nil || a.AI.CommandStatus != AIStatusPending || a.AI.approveBusy {
		return
	}
	cmd := strings.TrimSpace(a.AI.CommandDraft)
	if cmd == "" {
		cmd = a.AI.PendingCommand.Command
	}
	a.AI.approveBusy = true
	a.AI.CommandStatus = AIStatusApproved
	a.AI.CommandEditing = false
	a.saveAISnapshot()
	sid := a.AI.SessionID
	if a.Backend == nil {
		a.AI.approveBusy = false
		return
	}
	b := a.Backend
	go func() {
		err := b.ApproveAICommand(sid, cmd)
		a.Post(func() {
			a.AI.approveBusy = false
			if err != nil && sid == a.AI.SessionID {
				a.AI.CommandStatus = AIStatusPending
				a.AI.Err = "批准失败：" + err.Error()
				a.saveAISnapshot()
			}
		})
	}()
}

// RejectAICommand rejects the pending command.
func (a *App) RejectAICommand() {
	if a.AI.PendingCommand == nil || a.AI.CommandStatus != AIStatusPending || a.AI.approveBusy {
		return
	}
	a.AI.approveBusy = true
	a.AI.CommandStatus = AIStatusRejected
	a.saveAISnapshot()
	sid := a.AI.SessionID
	if a.Backend == nil {
		a.AI.approveBusy = false
		return
	}
	b := a.Backend
	go func() {
		err := b.RejectAICommand(sid)
		a.Post(func() {
			a.AI.approveBusy = false
			if err != nil && sid == a.AI.SessionID {
				a.AI.CommandStatus = AIStatusPending
				a.AI.Err = "拒绝失败：" + err.Error()
				a.saveAISnapshot()
			}
		})
	}()
}

// ApproveAIPlan approves the pending plan.
func (a *App) ApproveAIPlan() {
	if a.AI.PendingPlan == nil || a.AI.PlanStatus != AIStatusPending || a.AI.approveBusy {
		return
	}
	a.AI.approveBusy = true
	a.AI.PlanStatus = AIStatusApproved
	a.saveAISnapshot()
	sid := a.AI.SessionID
	if a.Backend == nil {
		a.AI.approveBusy = false
		return
	}
	b := a.Backend
	go func() {
		err := b.ApproveAIPlan(sid)
		a.Post(func() {
			a.AI.approveBusy = false
			if err != nil && sid == a.AI.SessionID {
				a.AI.PlanStatus = AIStatusPending
				a.AI.Err = "批准计划失败：" + err.Error()
				a.saveAISnapshot()
			}
		})
	}()
}

// RejectAIPlan rejects the pending plan.
func (a *App) RejectAIPlan() {
	if a.AI.PendingPlan == nil || a.AI.PlanStatus != AIStatusPending || a.AI.approveBusy {
		return
	}
	a.AI.approveBusy = true
	a.AI.PlanStatus = AIStatusRejected
	a.saveAISnapshot()
	sid := a.AI.SessionID
	if a.Backend == nil {
		a.AI.approveBusy = false
		return
	}
	b := a.Backend
	go func() {
		err := b.RejectAIPlan(sid)
		a.Post(func() {
			a.AI.approveBusy = false
			if err != nil && sid == a.AI.SessionID {
				a.AI.PlanStatus = AIStatusPending
				a.AI.Err = "拒绝计划失败：" + err.Error()
				a.saveAISnapshot()
			}
		})
	}()
}

// CancelAIRun cancels the running turn.
func (a *App) CancelAIRun() {
	sid := a.AI.SessionID
	if sid == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		if err := b.CancelAIRun(sid); err != nil {
			a.Post(func() {
				if sid == a.AI.SessionID {
					a.AI.Err = "取消失败：" + err.Error()
				}
			})
		}
	}()
}

// ClearAIConversation removes all messages of the active conversation.
func (a *App) ClearAIConversation() {
	sid := a.AI.SessionID
	if sid == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.ClearAIMessages(sid)
		a.Post(func() {
			if sid != a.AI.SessionID {
				return
			}
			if err != nil {
				a.AI.Err = "清空失败：" + err.Error()
				return
			}
			a.AI.Messages = nil
			a.AI.Streaming = ""
			a.AI.PendingCommand = nil
			a.AI.CommandStatus = ""
			a.AI.CommandDraft = ""
			a.AI.PendingPlan = nil
			a.AI.PlanStatus = ""
			a.AI.Err = ""
			a.AI.RunningCommand = ""
			a.AI.RunOutput = ""
			a.saveAISnapshot()
		})
	}()
}

// NewAIConversation creates and switches to a fresh conversation.
func (a *App) NewAIConversation() {
	a.saveAISnapshot()
	host := a.AI.HostID
	if host == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		sid, err := b.NewAISession(host, "AI 对话")
		convs, _ := b.ListAIConversations(host)
		var prov configstore.AIProvider
		var provErr error
		if err == nil {
			prov, provErr = b.ProviderForConversation(sid)
		}
		a.Post(func() {
			if host != a.AI.HostID {
				return
			}
			if err != nil {
				a.AI.Err = "新建对话失败：" + err.Error()
				return
			}
			a.AI.providerSeq++
			a.applyAIProvider(prov, provErr == nil)
			a.AI.SessionID = sid
			a.AI.Messages = nil
			a.AI.Streaming = ""
			a.AI.PendingCommand = nil
			a.AI.CommandStatus = ""
			a.AI.CommandDraft = ""
			a.AI.PendingPlan = nil
			a.AI.PlanStatus = ""
			a.AI.State = ""
			a.AI.Err = ""
			a.AI.RunningCommand = ""
			a.AI.RunOutput = ""
			if convs != nil {
				a.AI.Conversations = convs
			}
			a.AI.HistoryOpen = false
		})
	}()
}

// SwitchAIConversation opens a conversation from the history list.
func (a *App) SwitchAIConversation(sid string) {
	if sid == "" || sid == a.AI.SessionID {
		a.AI.HistoryOpen = false
		return
	}
	a.saveAISnapshot()
	a.AI.SessionID = sid
	a.AI.Messages = nil
	a.refreshAIProviderContext()
	a.restoreAISnapshot(sid)
	a.AI.HistoryOpen = false
	a.AI.RenameID = ""
	a.AI.DeleteConfirm = ""
	a.resyncAI(sid)
}

// DeleteAIConversation deletes a conversation; if it is the active one,
// the newest remaining conversation is attached instead.
func (a *App) DeleteAIConversation(sid string) {
	host := a.AI.HostID
	if sid == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.DeleteAIConversation(sid)
		a.Post(func() {
			if host != a.AI.HostID {
				return
			}
			if err != nil {
				a.AI.Err = "删除对话失败：" + err.Error()
				return
			}
			a.AI.DeleteConfirm = ""
			delete(a.AI.snapshots, sid)
			if sid == a.AI.SessionID {
				// Clear first so attachAI's save cannot resurrect the
				// snapshot of the conversation just deleted.
				a.AI.SessionID = ""
				a.attachAI(host)
				return
			}
			a.refreshAIConversations(host)
		})
	}()
}

// refreshAIConversations reloads the history list without touching the
// active conversation.
func (a *App) refreshAIConversations(host string) {
	if a.Backend == nil {
		return
	}
	b := a.Backend
	seq := a.AI.attachSeq
	go func() {
		convs, err := b.ListAIConversations(host)
		if err != nil {
			return
		}
		a.Post(func() {
			if host != a.AI.HostID || seq != a.AI.attachSeq {
				return
			}
			a.AI.Conversations = convs
		})
	}()
}

// RenameAIConversation changes a conversation title.
func (a *App) RenameAIConversation(sid, title string) {
	host := a.AI.HostID
	title = strings.TrimSpace(title)
	if sid == "" || title == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.RenameAIConversation(sid, title)
		convs, _ := b.ListAIConversations(host)
		a.Post(func() {
			if host != a.AI.HostID {
				return
			}
			if err != nil {
				a.AI.Err = "重命名失败：" + err.Error()
				return
			}
			a.AI.RenameID = ""
			a.AI.RenameTitle = ""
			if convs != nil {
				a.AI.Conversations = convs
			}
		})
	}()
}

// ToggleToolOutput folds or unfolds a tool output block.
func (a *App) ToggleToolOutput(messageID string) {
	if a.AI.ToolOpen == nil {
		a.AI.ToolOpen = map[string]bool{}
	}
	a.AI.ToolOpen[messageID] = !a.AI.ToolOpen[messageID]
}
