// Package audit exposes AI call audit records.
package audit

import logsstore "nextopsx/store/logs"

// Service exposes AI call audit records.
type Service struct {
	logs *logsstore.LogsStore
}

// New builds the audit service.
func New(logs *logsstore.LogsStore) *Service {
	return &Service{logs: logs}
}

// ListAICallLogs returns recent AI call audit records (newest first).
func (s *Service) ListAICallLogs(limit int) ([]logsstore.CallLog, error) {
	return s.logs.ListLogs(limit)
}

// AITokenSummary returns the token usage aggregate over all audit records.
func (s *Service) AITokenSummary() (logsstore.TokenSummary, error) {
	return s.logs.TokenSummary()
}
