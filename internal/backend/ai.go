package backend

import (
	configstore "nextopsx/internal/store/config"
	convstore "nextopsx/internal/store/conversations"
)

// The methods below are the native AI panel's facade over the session and
// configuration handlers. They stay framework-independent and never touch
// the UI; callers run them off the UI thread and republish results.

// EnsureAISession returns the asset's current conversation, creating one
// when it has none.
func (b *Backend) EnsureAISession(hostID string) (string, error) {
	return b.sessions.EnsureSession(hostID)
}

// NewAISession creates a conversation (the panel's "new chat" action).
func (b *Backend) NewAISession(hostID, title string) (string, error) {
	return b.sessions.NewSession(hostID, title)
}

// ListAIConversations lists an asset's conversations, newest first.
func (b *Backend) ListAIConversations(hostID string) ([]convstore.Conversation, error) {
	return b.sessions.ListConversations(hostID)
}

// LoadAIMessages loads a conversation's persisted messages.
func (b *Backend) LoadAIMessages(sessionID string) ([]convstore.Message, error) {
	return b.sessions.LoadMessages(sessionID)
}

// SendAIMessage starts one agent turn.
func (b *Backend) SendAIMessage(sessionID, text string) error {
	return b.sessions.SendMessage(sessionID, text)
}

// ApproveAICommand approves the pending command, possibly edited by the user.
func (b *Backend) ApproveAICommand(sessionID, command string) error {
	return b.sessions.ApproveCommand(sessionID, command)
}

// RejectAICommand rejects the pending command and asks the model for another approach.
func (b *Backend) RejectAICommand(sessionID string) error {
	return b.sessions.RejectCommand(sessionID)
}

// ApproveAIPlan approves the pending execution plan.
func (b *Backend) ApproveAIPlan(sessionID string) error {
	return b.sessions.ApprovePlan(sessionID)
}

// RejectAIPlan rejects the pending execution plan.
func (b *Backend) RejectAIPlan(sessionID string) error {
	return b.sessions.RejectPlan(sessionID)
}

// CancelAIRun cancels the running turn.
func (b *Backend) CancelAIRun(sessionID string) error {
	return b.sessions.CancelRun(sessionID)
}

// ClearAIMessages removes a conversation's messages but keeps the conversation.
func (b *Backend) ClearAIMessages(sessionID string) error {
	return b.sessions.ClearMessages(sessionID)
}

// DeleteAIConversation deletes a conversation and its runtime state.
func (b *Backend) DeleteAIConversation(sessionID string) error {
	return b.sessions.DeleteConversation(sessionID)
}

// RenameAIConversation changes a conversation's title.
func (b *Backend) RenameAIConversation(sessionID, title string) error {
	return b.sessions.RenameConversation(sessionID, title)
}

// GetAIConfig returns the model configuration (API key decrypted).
func (b *Backend) GetAIConfig() (configstore.AIConfig, error) {
	return b.aiConfig.GetAIConfig()
}

// SaveAIConfig stores the model configuration and invalidates cached models.
func (b *Backend) SaveAIConfig(c configstore.AIConfig) error {
	return b.aiConfig.SaveAIConfig(c)
}
