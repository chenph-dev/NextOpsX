package backend

import (
	"nextopsx/internal/store"
	configstore "nextopsx/internal/store/config"
)

// GetUISettings returns the persisted UI preferences (theme, …).
func (b *Backend) GetUISettings() (configstore.UISettings, error) {
	return b.uiSettings.GetUISettings()
}

// SaveUISettings stores UI preferences.
func (b *Backend) SaveUISettings(u configstore.UISettings) error {
	return b.uiSettings.SaveUISettings(u)
}

// DataDir returns the application's data directory, for display in settings.
func (b *Backend) DataDir() string {
	dir, err := store.DataDir()
	if err != nil {
		return ""
	}
	return dir
}
