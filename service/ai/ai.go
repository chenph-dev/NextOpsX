// Package ai exposes AI sessions and provider configuration.
package ai

import (
	"nextopsx/agent/session"
	configstore "nextopsx/store/config"
	convstore "nextopsx/store/conversations"
)

// Service exposes AI sessions and provider configuration.
type Service struct {
	sessions  *sessionHandler
	providers *providerHandler
	sm        *session.SessionManager
}

// New builds the AI service. onChange is called after provider changes.
func New(sm *session.SessionManager, convs *convstore.ConvStore, providers *configstore.ProviderStore, onChange func()) *Service {
	return &Service{
		sessions:  newSessionHandler(convs, sm),
		providers: newProviderHandler(providers, onChange),
		sm:        sm,
	}
}

// The methods below are the native AI panel's facade over the session and
// configuration handlers. They stay framework-independent and never touch
// the UI; callers run them off the UI thread and republish results.

// EnsureAISession returns the asset's current conversation, creating one
// when it has none.
func (s *Service) EnsureAISession(hostID string) (string, error) {
	return s.sessions.EnsureSession(hostID)
}

// NewAISession creates a conversation (the panel's "new chat" action).
func (s *Service) NewAISession(hostID, title string) (string, error) {
	return s.sessions.NewSession(hostID, title)
}

// ListAIConversations lists an asset's conversations, newest first.
func (s *Service) ListAIConversations(hostID string) ([]convstore.Conversation, error) {
	return s.sessions.ListConversations(hostID)
}

// LoadAIMessages loads a conversation's persisted messages.
func (s *Service) LoadAIMessages(sessionID string) ([]convstore.Message, error) {
	return s.sessions.LoadMessages(sessionID)
}

// SendAIMessage starts one agent turn.
func (s *Service) SendAIMessage(sessionID, text string) error {
	return s.sessions.SendMessage(sessionID, text)
}

// ApproveAICommand approves the pending command, possibly edited by the user.
func (s *Service) ApproveAICommand(sessionID, command string) error {
	return s.sessions.ApproveCommand(sessionID, command)
}

// RejectAICommand rejects the pending command and asks the model for another approach.
func (s *Service) RejectAICommand(sessionID string) error {
	return s.sessions.RejectCommand(sessionID)
}

// ApproveAIPlan approves the pending execution plan.
func (s *Service) ApproveAIPlan(sessionID string) error {
	return s.sessions.ApprovePlan(sessionID)
}

// RejectAIPlan rejects the pending execution plan.
func (s *Service) RejectAIPlan(sessionID string) error {
	return s.sessions.RejectPlan(sessionID)
}

// CancelAIRun cancels the running turn.
func (s *Service) CancelAIRun(sessionID string) error {
	return s.sessions.CancelRun(sessionID)
}

// ClearAIMessages removes a conversation's messages but keeps the conversation.
func (s *Service) ClearAIMessages(sessionID string) error {
	return s.sessions.ClearMessages(sessionID)
}

// DeleteAIConversation deletes a conversation and its runtime state.
func (s *Service) DeleteAIConversation(sessionID string) error {
	return s.sessions.DeleteConversation(sessionID)
}

// RenameAIConversation changes a conversation's title.
func (s *Service) RenameAIConversation(sessionID, title string) error {
	return s.sessions.RenameConversation(sessionID, title)
}

// ListAIProviders returns all providers without API keys.
func (s *Service) ListAIProviders() ([]configstore.AIProvider, error) {
	return s.providers.List()
}

// GetAIProvider returns one provider with its API key decrypted (edit form).
func (s *Service) GetAIProvider(id string) (configstore.AIProvider, error) {
	return s.providers.Get(id)
}

// CreateAIProvider adds a provider and returns its ID.
func (s *Service) CreateAIProvider(p configstore.AIProvider) (string, error) {
	return s.providers.Create(p)
}

// UpdateAIProvider updates a provider.
func (s *Service) UpdateAIProvider(p configstore.AIProvider) error {
	return s.providers.Update(p)
}

// DeleteAIProvider removes a provider; conversations fall back to the default.
func (s *Service) DeleteAIProvider(id string) error {
	return s.providers.Delete(id)
}

// SetDefaultAIProvider marks a provider as the default for new conversations.
func (s *Service) SetDefaultAIProvider(id string) error {
	return s.providers.SetDefault(id)
}

// SetConversationAIProvider selects the provider used by one conversation.
func (s *Service) SetConversationAIProvider(sessionID, providerID string) error {
	return s.sm.SetConversationProvider(sessionID, providerID)
}

// ProviderForConversation returns the provider a conversation resolves to.
func (s *Service) ProviderForConversation(sessionID string) (configstore.AIProvider, error) {
	return s.sm.ProviderForConversation(sessionID)
}
