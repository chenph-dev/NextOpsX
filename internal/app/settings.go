package app

import (
	"strings"

	configstore "nextopsx/internal/store/config"
)

// MainView is the window's top-level view.
type MainView int

const (
	// ViewAssets is the asset tree plus the workbench/AI panel.
	ViewAssets MainView = iota
	// ViewSettings is the full-window settings page.
	ViewSettings
)

// SettingsTab identifies a section of the settings page.
type SettingsTab int

const (
	SettingsModel SettingsTab = iota
	SettingsSkills
	SettingsPolicy
	SettingsAudit
)

// SettingsState is the global settings page.
type SettingsState struct {
	Tab    SettingsTab
	Notice string

	// 模型分区：供应商列表与编辑草稿。
	Providers        []configstore.AIProvider
	ProvidersLoading bool
	ProvidersErr     string
	SelectedProvider string
	Draft            configstore.AIProvider
	DraftIsNew       bool
	DraftLoading     bool
	DraftSaving      bool
	ConfirmDelete    string // 待确认删除的供应商 ID

	providerSeq uint64
	draftSeq    uint64
}

// OpenSettings switches the window to the settings page and selects tab.
func (a *App) OpenSettings(tab SettingsTab) {
	a.View = ViewSettings
	a.Settings.Tab = tab
	a.Settings.Notice = ""
	switch tab {
	case SettingsModel:
		a.LoadAIProviders()
	case SettingsSkills:
		a.LoadSkills()
	case SettingsPolicy:
		a.LoadApprovalPolicy()
	case SettingsAudit:
		a.LoadAudit()
	}
}

// CloseSettings returns to the asset view.
func (a *App) CloseSettings() {
	a.View = ViewAssets
	a.Settings.Notice = ""
}

// LoadAIProviders refreshes the provider list.
func (a *App) LoadAIProviders() {
	a.Settings.providerSeq++
	seq := a.Settings.providerSeq
	a.Settings.ProvidersLoading = true
	a.Settings.ProvidersErr = ""
	if a.Backend == nil {
		a.Settings.ProvidersLoading = false
		a.Settings.ProvidersErr = "后端未初始化"
		return
	}
	b := a.Backend
	go func() {
		list, err := b.ListAIProviders()
		a.Post(func() {
			if seq != a.Settings.providerSeq {
				return
			}
			a.Settings.ProvidersLoading = false
			if err != nil {
				a.Settings.ProvidersErr = "读取供应商失败：" + err.Error()
				return
			}
			a.Settings.Providers = list
			if a.Settings.SelectedProvider != "" && !providerExists(list, a.Settings.SelectedProvider) {
				a.Settings.SelectedProvider = ""
				a.Settings.Draft = configstore.AIProvider{}
				a.Settings.DraftIsNew = false
			}
			if a.Settings.SelectedProvider == "" && !a.Settings.DraftIsNew && len(list) > 0 {
				a.SelectAIProvider(defaultProviderID(list))
			}
		})
	}()
}

// SelectAIProvider loads one provider into the editor.
func (a *App) SelectAIProvider(id string) {
	if id == "" {
		return
	}
	if a.Backend == nil {
		a.Settings.ProvidersErr = "后端未初始化"
		return
	}
	a.Settings.draftSeq++
	seq := a.Settings.draftSeq
	a.Settings.SelectedProvider = id
	a.Settings.DraftIsNew = false
	a.Settings.DraftLoading = true
	a.Settings.ProvidersErr = ""
	b := a.Backend
	go func() {
		p, err := b.GetAIProvider(id)
		a.Post(func() {
			if seq != a.Settings.draftSeq || a.Settings.SelectedProvider != id {
				return
			}
			a.Settings.DraftLoading = false
			if err != nil {
				a.Settings.ProvidersErr = "读取供应商失败：" + err.Error()
				// 清掉旧草稿，避免把另一个供应商的数据保存到当前选中项。
				a.Settings.Draft = configstore.AIProvider{ID: id}
				return
			}
			a.Settings.Draft = p
		})
	}()
}

// NewAIProvider starts an empty provider draft.
func (a *App) NewAIProvider() {
	a.Settings.draftSeq++
	a.Settings.SelectedProvider = ""
	a.Settings.DraftIsNew = true
	a.Settings.DraftLoading = false
	a.Settings.DraftSaving = false
	a.Settings.ProvidersErr = ""
	a.Settings.Draft = configstore.AIProvider{Protocol: "openai"}
}

// CancelNewAIProvider discards an unsaved new-provider draft.
func (a *App) CancelNewAIProvider() {
	a.Settings.draftSeq++
	a.Settings.DraftIsNew = false
	a.Settings.Draft = configstore.AIProvider{}
	a.Settings.ProvidersErr = ""
}

// SaveAIProvider creates or updates the provider draft.
func (a *App) SaveAIProvider() {
	d := a.Settings.Draft
	d.Name = strings.TrimSpace(d.Name)
	d.Protocol = strings.TrimSpace(d.Protocol)
	d.Model = strings.TrimSpace(d.Model)
	d.BaseURL = strings.TrimSpace(d.BaseURL)
	d.APIKey = strings.TrimSpace(d.APIKey)
	a.Settings.Notice = ""
	if d.Name == "" || d.Protocol == "" || d.Model == "" {
		a.Settings.ProvidersErr = "请填写名称、协议与模型"
		return
	}
	if a.Backend == nil {
		a.Settings.ProvidersErr = "后端未初始化"
		return
	}
	a.Settings.ProvidersErr = ""
	a.Settings.DraftSaving = true
	isNew := a.Settings.DraftIsNew
	if !isNew && (d.ID == "" || d.ID != a.Settings.SelectedProvider) {
		a.Settings.DraftSaving = false
		a.Settings.ProvidersErr = "供应商尚未加载完成"
		return
	}
	// 保存期间用户可能切换/新建了草稿；用序号判断是否还应回填编辑区。
	draftSeq := a.Settings.draftSeq
	b := a.Backend
	go func() {
		id := d.ID
		var err error
		if isNew {
			id, err = b.CreateAIProvider(d)
		} else {
			err = b.UpdateAIProvider(d)
		}
		a.Post(func() {
			a.Settings.DraftSaving = false
			if err != nil {
				a.Settings.ProvidersErr = "保存供应商失败：" + err.Error()
				return
			}
			a.Settings.Notice = "供应商已保存"
			a.LoadAIProviders()
			if draftSeq == a.Settings.draftSeq {
				a.Settings.DraftIsNew = false
				a.Settings.SelectedProvider = id
				a.SelectAIProvider(id)
			}
			a.refreshAIProviderContext()
		})
	}()
}

// OpenDeleteAIProvider asks for confirmation before deleting a provider.
func (a *App) OpenDeleteAIProvider(id string) {
	if id != "" {
		a.Settings.ConfirmDelete = id
	}
}

// ConfirmDeleteAIProvider deletes the confirmed provider.
func (a *App) ConfirmDeleteAIProvider() {
	id := a.Settings.ConfirmDelete
	if id == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.DeleteAIProvider(id)
		a.Post(func() {
			a.Settings.ConfirmDelete = ""
			if err != nil {
				a.Settings.ProvidersErr = "删除供应商失败：" + err.Error()
				return
			}
			a.Settings.Notice = "供应商已删除"
			if a.Settings.SelectedProvider == id {
				a.Settings.SelectedProvider = ""
				a.Settings.Draft = configstore.AIProvider{}
				a.Settings.DraftIsNew = false
			}
			a.LoadAIProviders()
			a.refreshAIProviderContext()
		})
	}()
}

// SetDefaultAIProvider marks id as the default for new conversations.
func (a *App) SetDefaultAIProvider(id string) {
	if id == "" || a.Backend == nil {
		return
	}
	b := a.Backend
	go func() {
		err := b.SetDefaultAIProvider(id)
		a.Post(func() {
			if err != nil {
				a.Settings.ProvidersErr = "设置默认供应商失败：" + err.Error()
				return
			}
			a.Settings.Notice = "已设为默认供应商"
			a.LoadAIProviders()
			if a.Settings.SelectedProvider != "" {
				a.SelectAIProvider(a.Settings.SelectedProvider)
			}
			a.refreshAIProviderContext()
		})
	}()
}

// refreshAIProviderContext re-reads the current conversation's provider after
// the provider list changed (delete, default switch, model edit).
func (a *App) refreshAIProviderContext() {
	sid := a.AI.SessionID
	if sid == "" || a.Backend == nil {
		return
	}
	a.AI.providerSeq++
	seq := a.AI.providerSeq
	b := a.Backend
	go func() {
		prov, provErr := b.ProviderForConversation(sid)
		list, listErr := b.ListAIProviders()
		a.Post(func() {
			if a.AI.SessionID != sid || seq != a.AI.providerSeq {
				return
			}
			if listErr == nil {
				a.AI.Providers = list
			}
			// 用户正在切换供应商时，以用户的显式选择为准。
			if !a.AI.switching {
				a.applyAIProvider(prov, provErr == nil)
			}
		})
	}()
}

func providerExists(list []configstore.AIProvider, id string) bool {
	for _, p := range list {
		if p.ID == id {
			return true
		}
	}
	return false
}

func defaultProviderID(list []configstore.AIProvider) string {
	for _, p := range list {
		if p.IsDefault {
			return p.ID
		}
	}
	if len(list) > 0 {
		return list[0].ID
	}
	return ""
}
