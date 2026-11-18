package ai

import (
	"nextopsx/agent/session"
	convstore "nextopsx/store/conversations"
)

// sessionHandler 处理会话/对话相关的前端调用。
// 事件推送不经过本 handler：SessionManager 持有组合根注入的
// emitEvent 函数直接推送事件。
type sessionHandler struct {
	convs          *convstore.ConvStore
	sessionManager *session.SessionManager
}

// newSessionHandler 构造 sessionHandler。
func newSessionHandler(convs *convstore.ConvStore, sm *session.SessionManager) *sessionHandler {
	return &sessionHandler{convs: convs, sessionManager: sm}
}

// EnsureSession 获取（必要时懒创建）资产当前会话。
func (h *sessionHandler) EnsureSession(hostID string) (string, error) {
	return h.sessionManager.EnsureSession(hostID)
}

// NewSession 新建一个会话（前端"新建对话"）。
func (h *sessionHandler) NewSession(hostID, title string) (string, error) {
	return h.sessionManager.CreateSession(hostID, title)
}

// SendMessage 发送消息。executor 由 SessionManager 通过注入的
// ExecutorFor 工厂解析，避免把 sessionID 误当 hostID。
func (h *sessionHandler) SendMessage(sid, text string) error {
	return h.sessionManager.SendMessage(sid, text)
}

func (h *sessionHandler) ApproveCommand(sid, command string) error {
	return h.sessionManager.ApproveCommand(sid, command)
}

func (h *sessionHandler) RejectCommand(sid string) error {
	return h.sessionManager.RejectCommand(sid)
}

// ApprovePlan 批准执行计划（计划模式）。
func (h *sessionHandler) ApprovePlan(sid string) error {
	return h.sessionManager.ApprovePlan(sid)
}

// RejectPlan 拒绝执行计划（计划模式）。
func (h *sessionHandler) RejectPlan(sid string) error {
	return h.sessionManager.RejectPlan(sid)
}

func (h *sessionHandler) CancelRun(sid string) error {
	return h.sessionManager.CancelRun(sid)
}

// ClearMessages 清空会话全部消息（输入框快捷命令 /clear）。
func (h *sessionHandler) ClearMessages(sid string) error {
	return h.sessionManager.ClearMessages(sid)
}

// RenameConversation 重命名会话标题（历史菜单手动重命名）。
func (h *sessionHandler) RenameConversation(sid, title string) error {
	return h.convs.RenameConversation(sid, title)
}

func (h *sessionHandler) ListConversations(hostID string) ([]convstore.Conversation, error) {
	return h.convs.ListConversations(hostID)
}

func (h *sessionHandler) LoadMessages(sid string) ([]convstore.Message, error) {
	return h.convs.LoadMessages(sid)
}

func (h *sessionHandler) DeleteConversation(sid string) error {
	return h.sessionManager.DeleteSession(sid)
}
