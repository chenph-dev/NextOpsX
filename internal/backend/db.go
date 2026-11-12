package backend

import "nextopsx/internal/dbexec"

// ListSchema returns a database asset's tables, views and columns.
// Registry-only drivers (Redis, Elasticsearch) have no relational schema
// and return an empty schema.
func (b *Backend) ListSchema(hostID string) (*dbexec.Schema, error) {
	return b.dbHandler.ListSchema(hostID)
}

// ExecuteSQL runs one SQL statement or query on a database asset. Query
// statements return columns and rows; writes return the affected count.
// It blocks for up to 30s and must be called off the UI thread.
func (b *Backend) ExecuteSQL(hostID, sqlText string) (*dbexec.Result, error) {
	return b.dbHandler.ExecuteSQL(hostID, sqlText)
}
