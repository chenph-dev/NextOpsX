// Package configstore 管理 AI 供应商配置的读写与加密。
package configstore

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"

	"nextopsx/internal/store"
	"nextopsx/internal/store/crypto"
)

// providerModel 对应 ai_providers 表。
type providerModel struct {
	ID              string `gorm:"column:id;primaryKey"`
	Name            string `gorm:"column:name;uniqueIndex"`
	Protocol        string `gorm:"column:protocol"`
	Model           string `gorm:"column:model"`
	BaseURL         string `gorm:"column:base_url"`
	APIKeyEncrypted []byte `gorm:"column:api_key_encrypted"`
	IsDefault       bool   `gorm:"column:is_default"`
	CreatedAt       int64  `gorm:"column:created_at"`
	UpdatedAt       int64  `gorm:"column:updated_at"`
}

func (providerModel) TableName() string { return "ai_providers" }

// AIProvider 一条命名的 AI 供应商配置（DTO）。
// Protocol 是协议语义（openai 兼容 / claude），不是厂商名。
type AIProvider struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Model     string `json:"model"`
	BaseURL   string `json:"baseURL"`
	APIKey    string `json:"apiKey,omitempty"` // 仅单条读取时返回明文；落库加密
	IsDefault bool   `json:"isDefault"`
	CreatedAt int64  `json:"createdAt"`
	UpdatedAt int64  `json:"updatedAt"`
}

// ProviderStore 提供 AI 供应商 CRUD。
type ProviderStore struct {
	app *store.DB
}

// NewProviderStore 构造 ProviderStore。
func NewProviderStore(app *store.DB) *ProviderStore {
	return &ProviderStore{app: app}
}

// List 返回全部供应商（默认供应商在前，其次按名称），不含 API Key。
func (s *ProviderStore) List() ([]AIProvider, error) {
	var rows []providerModel
	if err := s.app.GORM().Order("is_default desc, name").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]AIProvider, 0, len(rows))
	for _, r := range rows {
		out = append(out, toProvider(r, ""))
	}
	return out, nil
}

// Get 按 ID 读取供应商（含解密后的 API Key）。
func (s *ProviderStore) Get(id string) (AIProvider, error) {
	var m providerModel
	if err := s.app.GORM().First(&m, "id = ?", id).Error; err != nil {
		return AIProvider{}, err
	}
	key, err := s.decrypt(m.APIKeyEncrypted)
	if err != nil {
		return AIProvider{}, err
	}
	return toProvider(m, key), nil
}

// Default 返回默认供应商（含 API Key）。没有标记默认时回落到最早创建的一条。
func (s *ProviderStore) Default() (AIProvider, error) {
	var m providerModel
	err := s.app.GORM().Where("is_default = ?", true).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		err = s.app.GORM().Order("created_at").First(&m).Error
	}
	if err != nil {
		return AIProvider{}, err
	}
	key, err := s.decrypt(m.APIKeyEncrypted)
	if err != nil {
		return AIProvider{}, err
	}
	return toProvider(m, key), nil
}

// Resolve 解析一个供应商引用：id 为空或对应记录已删除时回落到默认供应商。
func (s *ProviderStore) Resolve(id string) (AIProvider, error) {
	if id != "" {
		p, err := s.Get(id)
		if err == nil {
			return p, nil
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return AIProvider{}, err
		}
	}
	return s.Default()
}

// Create 新增供应商。第一条供应商自动成为默认供应商。
func (s *ProviderStore) Create(p AIProvider) (string, error) {
	p = normalizeProvider(p)
	if err := validateProvider(p); err != nil {
		return "", err
	}
	enc, err := s.encrypt(p.APIKey)
	if err != nil {
		return "", err
	}
	id := crypto.NewID()
	now := time.Now().Unix()
	err = s.app.GORM().Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&providerModel{}).Count(&count).Error; err != nil {
			return err
		}
		isDefault := p.IsDefault || count == 0
		if isDefault {
			if err := tx.Model(&providerModel{}).Where("is_default = ?", true).
				Update("is_default", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(&providerModel{
			ID: id, Name: p.Name, Protocol: p.Protocol, Model: p.Model,
			BaseURL: p.BaseURL, APIKeyEncrypted: enc, IsDefault: isDefault,
			CreatedAt: now, UpdatedAt: now,
		}).Error
	})
	if err != nil {
		return "", providerErr(err, p.Name)
	}
	return id, nil
}

// Update 更新供应商；默认供应商标记不受此方法影响（用 SetDefault 切换）。
func (s *ProviderStore) Update(p AIProvider) error {
	p = normalizeProvider(p)
	if err := validateProvider(p); err != nil {
		return err
	}
	enc, err := s.encrypt(p.APIKey)
	if err != nil {
		return err
	}
	var existing providerModel
	if err := s.app.GORM().First(&existing, "id = ?", p.ID).Error; err != nil {
		return err
	}
	err = s.app.GORM().Model(&providerModel{}).Where("id = ?", p.ID).Updates(map[string]any{
		"name":              p.Name,
		"protocol":          p.Protocol,
		"model":             p.Model,
		"base_url":          p.BaseURL,
		"api_key_encrypted": enc,
		"updated_at":        time.Now().Unix(),
	}).Error
	if err != nil {
		return providerErr(err, p.Name)
	}
	return nil
}

// Delete 删除供应商。至少保留一个；删除默认供应商时自动把最早的一条设为默认；
// 引用它的历史对话回落为默认供应商（provider_id 置空）。
func (s *ProviderStore) Delete(id string) error {
	return s.app.GORM().Transaction(func(tx *gorm.DB) error {
		var count int64
		if err := tx.Model(&providerModel{}).Count(&count).Error; err != nil {
			return err
		}
		if count <= 1 {
			return fmt.Errorf("至少保留一个供应商")
		}
		var m providerModel
		if err := tx.First(&m, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Delete(&providerModel{}, "id = ?", id).Error; err != nil {
			return err
		}
		if m.IsDefault {
			var next providerModel
			if err := tx.Order("created_at").First(&next).Error; err != nil {
				return err
			}
			if err := tx.Model(&providerModel{}).Where("id = ?", next.ID).
				Update("is_default", true).Error; err != nil {
				return err
			}
		}
		return tx.Table("conversations").Where("provider_id = ?", id).
			Update("provider_id", "").Error
	})
}

// SetDefault 把 id 设为唯一默认供应商。
func (s *ProviderStore) SetDefault(id string) error {
	return s.app.GORM().Transaction(func(tx *gorm.DB) error {
		var m providerModel
		if err := tx.First(&m, "id = ?", id).Error; err != nil {
			return err
		}
		if err := tx.Model(&providerModel{}).Where("is_default = ?", true).
			Update("is_default", false).Error; err != nil {
			return err
		}
		return tx.Model(&providerModel{}).Where("id = ?", id).
			Update("is_default", true).Error
	})
}

func (s *ProviderStore) encrypt(key string) ([]byte, error) {
	if key == "" {
		return nil, nil
	}
	enc, err := s.app.Encrypt([]byte(key))
	if err != nil {
		return nil, fmt.Errorf("encrypt key: %w", err)
	}
	return enc, nil
}

func (s *ProviderStore) decrypt(blob []byte) (string, error) {
	if len(blob) == 0 {
		return "", nil
	}
	pt, err := s.app.Decrypt(blob)
	if err != nil {
		return "", fmt.Errorf("decrypt key: %w", err)
	}
	return string(pt), nil
}

func toProvider(m providerModel, key string) AIProvider {
	return AIProvider{
		ID: m.ID, Name: m.Name, Protocol: m.Protocol, Model: m.Model,
		BaseURL: m.BaseURL, APIKey: key, IsDefault: m.IsDefault,
		CreatedAt: m.CreatedAt, UpdatedAt: m.UpdatedAt,
	}
}

func normalizeProvider(p AIProvider) AIProvider {
	p.Name = strings.TrimSpace(p.Name)
	p.Protocol = strings.TrimSpace(p.Protocol)
	p.Model = strings.TrimSpace(p.Model)
	p.BaseURL = strings.TrimSpace(p.BaseURL)
	p.APIKey = strings.TrimSpace(p.APIKey)
	return p
}

func validateProvider(p AIProvider) error {
	if p.Name == "" || p.Protocol == "" || p.Model == "" {
		return fmt.Errorf("名称、协议与模型不能为空")
	}
	return nil
}

func providerErr(err error, name string) error {
	if strings.Contains(err.Error(), "UNIQUE constraint failed") {
		return fmt.Errorf("供应商名称 %q 已存在", name)
	}
	return err
}
