package backend

import logsstore "nextopsx/internal/store/logs"

// ListAICallLogs returns recent AI call audit records (newest first).
func (b *Backend) ListAICallLogs(limit int) ([]logsstore.CallLog, error) {
	return b.logsAPI.ListLogs(limit)
}

// AITokenSummary returns the token usage aggregate over all audit records.
func (b *Backend) AITokenSummary() (logsstore.TokenSummary, error) {
	return b.logsAPI.TokenSummary()
}
