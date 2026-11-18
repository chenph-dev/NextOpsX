package configstore

import (
	"slices"

	"nextopsx/i18n"
	"nextopsx/store"
)

// uiSettingsModel 界面偏好单行表（id=1）。
type uiSettingsModel struct {
	ID          int    `gorm:"column:id;primaryKey"`
	Theme       string `gorm:"column:theme"`
	Scale       int    `gorm:"column:scale"`
	Language    string `gorm:"column:language"`
	AIOpen      bool   `gorm:"column:ai_open"`
	AIHeight    int    `gorm:"column:ai_height"`
	SFTPOpen    bool   `gorm:"column:sftp_open"`
	SFTPWidth   int    `gorm:"column:sftp_width"`
	LastAssetID string `gorm:"column:last_asset_id"`
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

// 界面缩放的可选百分比。
const ScaleDefault = 100

// 工作台抽屉的默认尺寸与范围。
const (
	DefaultAIHeight  = 320
	MinAIHeight      = 200
	MaxAIHeight      = 480
	DefaultSFTPWidth = 330
	MinSFTPWidth     = 260
	MaxSFTPWidth     = 420
)

// NormalizeAIHeight clamps a persisted drawer height.
func NormalizeAIHeight(height int) int {
	if height <= 0 {
		return DefaultAIHeight
	}
	if height < MinAIHeight {
		return MinAIHeight
	}
	if height > MaxAIHeight {
		return MaxAIHeight
	}
	return height
}

// NormalizeSFTPWidth clamps a persisted drawer width.
func NormalizeSFTPWidth(width int) int {
	if width <= 0 {
		return DefaultSFTPWidth
	}
	if width < MinSFTPWidth {
		return MinSFTPWidth
	}
	if width > MaxSFTPWidth {
		return MaxSFTPWidth
	}
	return width
}

// Scales 是可选缩放比例的完整列表。
var Scales = []int{90, 100, 110, 125}

// NormalizeScale 把未知缩放比例回落为 100%。
func NormalizeScale(scale int) int {
	if slices.Contains(Scales, scale) {
		return scale
	}
	return ScaleDefault
}

// NormalizeLanguage 把未知语言回落为默认语言（zh-CN）。
func NormalizeLanguage(lang string) string {
	return string(i18n.Normalize(i18n.Lang(lang)))
}

// UISettings 界面偏好 DTO。Theme: "system" | "light" | "dark"；
// Scale 是界面缩放百分比；Language: "zh-CN" | "en-US"；其余字段保存
// 工作台抽屉的开合与尺寸。
type UISettings struct {
	Theme       string `json:"theme"`
	Scale       int    `json:"scale"`
	Language    string `json:"language"`
	AIOpen      bool   `json:"aiOpen"`
	AIHeight    int    `json:"aiHeight"`
	SFTPOpen    bool   `json:"sftpOpen"`
	SFTPWidth   int    `json:"sftpWidth"`
	LastAssetID string `json:"lastAssetId"`
}

// UISettingsStore 提供界面偏好读写。
type UISettingsStore struct {
	app *store.DB
}

// NewUISettingsStore 构造 UISettingsStore。
func NewUISettingsStore(app *store.DB) *UISettingsStore {
	return &UISettingsStore{app: app}
}

// GetUISettings 读取界面偏好；无记录或非法值回落到默认值。
func (s *UISettingsStore) GetUISettings() (UISettings, error) {
	// Find（而非 First）让「还没有记录」不产生 record-not-found 错误与日志。
	var m uiSettingsModel
	if err := s.app.GORM().Where("id = ?", 1).Limit(1).Find(&m).Error; err != nil {
		return UISettings{}, err
	}
	if m.ID == 0 {
		return UISettings{
			Theme: ThemeSystem, Scale: ScaleDefault, Language: string(i18n.Default),
			AIHeight: DefaultAIHeight, SFTPWidth: DefaultSFTPWidth,
		}, nil
	}
	return UISettings{
		Theme:       NormalizeTheme(m.Theme),
		Scale:       NormalizeScale(m.Scale),
		Language:    NormalizeLanguage(m.Language),
		AIOpen:      m.AIOpen,
		AIHeight:    NormalizeAIHeight(m.AIHeight),
		SFTPOpen:    m.SFTPOpen,
		SFTPWidth:   NormalizeSFTPWidth(m.SFTPWidth),
		LastAssetID: m.LastAssetID,
	}, nil
}

// SaveUISettings 保存界面偏好（覆盖单行 id=1）。
func (s *UISettingsStore) SaveUISettings(u UISettings) error {
	return s.app.GORM().Save(&uiSettingsModel{
		ID: 1, Theme: NormalizeTheme(u.Theme), Scale: NormalizeScale(u.Scale),
		Language: NormalizeLanguage(u.Language),
		AIOpen:   u.AIOpen, AIHeight: NormalizeAIHeight(u.AIHeight),
		SFTPOpen: u.SFTPOpen, SFTPWidth: NormalizeSFTPWidth(u.SFTPWidth),
		LastAssetID: u.LastAssetID,
	}).Error
}
