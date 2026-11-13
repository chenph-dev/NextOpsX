package configstore

import (
	"nextopsx/internal/store"
)

// uiSettingsModel 界面偏好单行表（id=1）。
type uiSettingsModel struct {
	ID    int    `gorm:"column:id;primaryKey"`
	Theme string `gorm:"column:theme"`
}

func (uiSettingsModel) TableName() string { return "ui_settings" }

// 主题取值；UI 与服务端共用这一份定义，避免新增主题时多处漂移。
const (
	ThemeSystem = "system"
	ThemeLight  = "light"
	ThemeDark   = "dark"
)

// Themes 是可选主题的完整列表。
var Themes = []string{ThemeSystem, ThemeLight, ThemeDark}

// NormalizeTheme 把未知主题回落为跟随系统。
func NormalizeTheme(theme string) string {
	switch theme {
	case ThemeLight, ThemeDark:
		return theme
	default:
		return ThemeSystem
	}
}

// UISettings 界面偏好 DTO。Theme: "system" | "light" | "dark"。
type UISettings struct {
	Theme string `json:"theme"`
}

// UISettingsStore 提供界面偏好读写。
type UISettingsStore struct {
	app *store.DB
}

// NewUISettingsStore 构造 UISettingsStore。
func NewUISettingsStore(app *store.DB) *UISettingsStore {
	return &UISettingsStore{app: app}
}

// GetUISettings 读取界面偏好；无记录或非法值回落到跟随系统。
func (s *UISettingsStore) GetUISettings() (UISettings, error) {
	// Find（而非 First）让「还没有记录」不产生 record-not-found 错误与日志。
	var m uiSettingsModel
	if err := s.app.GORM().Where("id = ?", 1).Limit(1).Find(&m).Error; err != nil {
		return UISettings{}, err
	}
	if m.ID == 0 {
		return UISettings{Theme: ThemeSystem}, nil
	}
	return UISettings{Theme: NormalizeTheme(m.Theme)}, nil
}

// SaveUISettings 保存界面偏好（覆盖单行 id=1）。
func (s *UISettingsStore) SaveUISettings(u UISettings) error {
	return s.app.GORM().Save(&uiSettingsModel{
		ID: 1, Theme: NormalizeTheme(u.Theme),
	}).Error
}
