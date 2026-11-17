// Package historystore persists commands executed by AI or from the app.
package historystore

import (
	"time"

	"nextopsx/internal/store"
	"nextopsx/internal/store/crypto"
)

// Entry is one executed command in the per-asset history.
type Entry struct {
	ID        string `gorm:"column:id;primaryKey" json:"id"`
	HostID    string `gorm:"column:host_id" json:"hostId"`
	SessionID string `gorm:"column:session_id" json:"sessionId"`
	Source    string `gorm:"column:source" json:"source"` // ai | snippet | app
	Command   string `gorm:"column:command" json:"command"`
	ExitCode  *int   `gorm:"column:exit_code" json:"exitCode,omitempty"`
	Output    string `gorm:"column:output" json:"output"`
	Ts        int64  `gorm:"column:ts" json:"ts"`
}

func (Entry) TableName() string { return "command_history" }

// Store provides command history reads and app-side writes.
type Store struct {
	app *store.DB
}

// NewStore constructs a history Store.
func NewStore(app *store.DB) *Store {
	return &Store{app: app}
}

// List returns the newest commands for a host.
func (s *Store) List(hostID string, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}
	var out []Entry
	if err := s.app.GORM().Where("host_id = ?", hostID).
		Order("ts DESC").Limit(limit).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Save inserts one command history entry, assigning an ID and timestamp when
// absent.
func (s *Store) Save(in Entry) error {
	if in.ID == "" {
		in.ID = crypto.NewID()
	}
	if in.Ts == 0 {
		in.Ts = time.Now().Unix()
	}
	return s.app.GORM().Create(&in).Error
}
