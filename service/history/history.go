// Package history records and lists executed commands.
package history

import historystore "nextopsx/store/history"

// Service records and lists executed commands.
type Service struct {
	store *historystore.Store
}

// New builds the command-history service.
func New(store *historystore.Store) *Service {
	return &Service{store: store}
}

// ListCommandHistory returns the newest executed commands for a host.
func (s *Service) ListCommandHistory(hostID string, limit int) ([]historystore.Entry, error) {
	return s.store.List(hostID, limit)
}

// SaveCommandHistory records a command executed from the app, such as a
// favorite or a snippet.
func (s *Service) SaveCommandHistory(hostID, source, command string, exitCode *int, output string) error {
	return s.store.Save(historystore.Entry{
		HostID: hostID, Source: source, Command: command,
		ExitCode: exitCode, Output: output,
	})
}
