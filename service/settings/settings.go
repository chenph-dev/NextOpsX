// Package settings stores UI preferences.
package settings

import (
	"nextopsx/store"
	configstore "nextopsx/store/config"
)

// Service stores UI preferences.
type Service struct {
	store *configstore.UISettingsStore
}

// New builds the settings service.
func New(store *configstore.UISettingsStore) *Service {
	return &Service{store: store}
}

// GetUISettings returns the persisted UI preferences (theme, …).
func (s *Service) GetUISettings() (configstore.UISettings, error) {
	return s.store.GetUISettings()
}

// SaveUISettings stores UI preferences.
func (s *Service) SaveUISettings(u configstore.UISettings) error {
	return s.store.SaveUISettings(u)
}

// DataDir returns the application's data directory, for display in settings.
func (s *Service) DataDir() string {
	dir, err := store.DataDir()
	if err != nil {
		return ""
	}
	return dir
}
