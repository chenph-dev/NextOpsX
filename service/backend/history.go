package backend

import historystore "nextopsx/store/history"

// ListCommandHistory returns the newest executed commands for a host.
func (b *Backend) ListCommandHistory(hostID string, limit int) ([]historystore.Entry, error) {
	return b.history.List(hostID, limit)
}

// SaveCommandHistory records a command executed from the app, such as a
// favorite or a snippet.
func (b *Backend) SaveCommandHistory(hostID, source, command string, exitCode *int, output string) error {
	return b.history.Save(historystore.Entry{
		HostID: hostID, Source: source, Command: command,
		ExitCode: exitCode, Output: output,
	})
}
