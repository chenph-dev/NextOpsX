package session

import (
	"errors"
	"fmt"

	"gorm.io/gorm"

	"nextopsx/internal/einoagent/checkpoint"
	agenttools "nextopsx/internal/einoagent/tools"
	configstore "nextopsx/internal/store/config"
)

// EnsureSession 返回资产当前会话（最近更新的 conversation），没有则懒创建。
func (m *SessionManager) EnsureSession(hostID string) (string, error) {
	convs, err := m.convs.ListConversations(hostID)
	if err != nil {
		return "", err
	}
	if len(convs) > 0 {
		return convs[0].ID, nil
	}
	return m.CreateSession(hostID, "AI 对话")
}

// CreateSession 新建一个 conversation（前端"新建对话"按钮），
// 默认继承当前默认供应商；未配置供应商时留空，发送时再报错。
func (m *SessionManager) CreateSession(hostID, title string) (string, error) {
	providerID := ""
	p, err := m.providers.Default()
	switch {
	case err == nil:
		providerID = p.ID
	case errors.Is(err, gorm.ErrRecordNotFound):
		// 还没有配置供应商：允许建会话，首次发送时再提示。
	default:
		return "", fmt.Errorf("读取默认供应商失败: %w", err)
	}
	return m.convs.NewConversationWithProvider(hostID, title, providerID)
}

// SetConversationProvider 指定会话使用的供应商（空串表示回落到默认），
// 并使已构建的 Graph 失效，下一轮按新供应商重建。
func (m *SessionManager) SetConversationProvider(sid, providerID string) error {
	if providerID != "" {
		if _, err := m.providers.Get(providerID); err != nil {
			return err
		}
	}
	if err := m.convs.SetConversationProvider(sid, providerID); err != nil {
		return err
	}
	m.InvalidateConfig()
	return nil
}

// ProviderForConversation 返回会话当前解析到的供应商（含 API Key）。
func (m *SessionManager) ProviderForConversation(sid string) (configstore.AIProvider, error) {
	conv, err := m.convs.GetConversation(sid)
	if err != nil {
		return configstore.AIProvider{}, err
	}
	return m.providers.Resolve(conv.ProviderID)
}

// sessionFor 获取或懒构建会话运行时。
func (m *SessionManager) sessionFor(sid string) (*agentSession, error) {
	m.mu.Lock()
	s, ok := m.sessions[sid]
	m.mu.Unlock()
	if ok {
		return s, nil
	}
	conv, err := m.convs.GetConversation(sid)
	if err != nil {
		return nil, fmt.Errorf("session %s not found: %w", sid, err)
	}
	s = &agentSession{
		id: sid, hostID: conv.HostID,
		checkpoints: checkpoint.NewMemCheckpointStore(),
		holder:      &executorHolder{},
		toolCalls:   agenttools.NewToolCallHolder(),
	}
	m.mu.Lock()
	if existing, ok := m.sessions[sid]; ok {
		s = existing
	} else {
		m.sessions[sid] = s
	}
	m.mu.Unlock()
	return s, nil
}

// sessionState 供测试与调试查询当前状态字符串。
func (m *SessionManager) sessionState(sid string) string {
	s, err := m.sessionFor(sid)
	if err != nil {
		return StateIdle
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.state.String()
}

// emitState 推送会话状态事件。
func (m *SessionManager) emitState(sid, state string) {
	if m.emit != nil {
		m.emit(sid, "session:state", state)
	}
}

// emitError 推送错误事件。cancelled 标记「主动取消」这一正常反馈，
// 前端据此用中性色展示而非红色告警（避免依赖错误文本做匹配）。
func (m *SessionManager) emitError(sid, message string, cancelled bool) {
	if m.emit != nil {
		m.emit(sid, "ai:error", map[string]any{"message": message, "cancelled": cancelled})
	}
}
