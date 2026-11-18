// Package snippetsstore 管理可复用的命令片段。
package snippetsstore

import (
	"fmt"
	"strings"
	"time"

	"nextopsx/store"
	"nextopsx/store/crypto"
)

// Snippet is one reusable command. Group is a free-form label used to
// organize the list; Command may contain {{name}}, {{host}}, {{addr}},
// {{user}}, {{port}} and {{protocol}} placeholders.
type Snippet struct {
	ID          string `gorm:"column:id;primaryKey" json:"id"`
	Name        string `gorm:"column:name" json:"name"`
	Group       string `gorm:"column:group_name" json:"group"`
	Command     string `gorm:"column:command" json:"command"`
	Description string `gorm:"column:description" json:"description"`
	CreatedAt   int64  `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt   int64  `gorm:"column:updated_at" json:"updatedAt"`
}

func (Snippet) TableName() string { return "command_snippets" }

// SnippetStore provides snippet CRUD.
type SnippetStore struct {
	app *store.DB
}

// NewSnippetStore constructs a SnippetStore.
func NewSnippetStore(app *store.DB) *SnippetStore {
	return &SnippetStore{app: app}
}

// List returns all snippets ordered by group and name.
func (s *SnippetStore) List() ([]Snippet, error) {
	var out []Snippet
	if err := s.app.GORM().Order("group_name, name").Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// Save inserts or updates a snippet. Empty ID inserts and assigns a new ID;
// otherwise the existing row is updated.
func (s *SnippetStore) Save(in Snippet) (Snippet, error) {
	in.Name = strings.TrimSpace(in.Name)
	in.Group = strings.TrimSpace(in.Group)
	in.Command = strings.TrimSpace(in.Command)
	in.Description = strings.TrimSpace(in.Description)
	if in.Name == "" {
		return Snippet{}, fmt.Errorf("片段名称不能为空")
	}
	if in.Command == "" {
		return Snippet{}, fmt.Errorf("命令不能为空")
	}
	now := time.Now().Unix()
	if in.ID == "" {
		in.ID = crypto.NewID()
		in.CreatedAt = now
		in.UpdatedAt = now
		if err := s.app.GORM().Create(&in).Error; err != nil {
			return Snippet{}, err
		}
		return in, nil
	}
	res := s.app.GORM().Model(&Snippet{}).Where("id = ?", in.ID).Updates(map[string]any{
		"name":        in.Name,
		"group_name":  in.Group,
		"command":     in.Command,
		"description": in.Description,
		"updated_at":  now,
	})
	if res.Error != nil {
		return Snippet{}, res.Error
	}
	if res.RowsAffected == 0 {
		return Snippet{}, fmt.Errorf("片段不存在")
	}
	var stored Snippet
	if err := s.app.GORM().First(&stored, "id = ?", in.ID).Error; err != nil {
		return Snippet{}, err
	}
	return stored, nil
}

// Delete removes a snippet.
func (s *SnippetStore) Delete(id string) error {
	res := s.app.GORM().Delete(&Snippet{}, "id = ?", id)
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("片段不存在")
	}
	return nil
}
