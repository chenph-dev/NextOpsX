// Package aiconfig 处理 AI 供应商配置相关的前端调用。
package aiconfig

import configstore "nextopsx/internal/store/config"

// ProviderHandler 处理 AI 供应商的 CRUD；任何变更都会通知会话管理器失效模型缓存。
type ProviderHandler struct {
	providers *configstore.ProviderStore
	onChange  func()
}

// NewProviderHandler 构造 ProviderHandler。onChange 可为 nil。
func NewProviderHandler(providers *configstore.ProviderStore, onChange func()) *ProviderHandler {
	return &ProviderHandler{providers: providers, onChange: onChange}
}

// List 返回全部供应商（不含 API Key）。
func (h *ProviderHandler) List() ([]configstore.AIProvider, error) {
	return h.providers.List()
}

// Get 返回单个供应商（含 API Key，供编辑表单使用）。
func (h *ProviderHandler) Get(id string) (configstore.AIProvider, error) {
	return h.providers.Get(id)
}

// Create 新增供应商，返回新 ID。
func (h *ProviderHandler) Create(p configstore.AIProvider) (string, error) {
	id, err := h.providers.Create(p)
	if err != nil {
		return "", err
	}
	h.changed()
	return id, nil
}

// Update 更新供应商。
func (h *ProviderHandler) Update(p configstore.AIProvider) error {
	if err := h.providers.Update(p); err != nil {
		return err
	}
	h.changed()
	return nil
}

// Delete 删除供应商；引用它的历史对话回落默认供应商。
func (h *ProviderHandler) Delete(id string) error {
	if err := h.providers.Delete(id); err != nil {
		return err
	}
	h.changed()
	return nil
}

// SetDefault 把 id 设为默认供应商。
func (h *ProviderHandler) SetDefault(id string) error {
	if err := h.providers.SetDefault(id); err != nil {
		return err
	}
	h.changed()
	return nil
}

func (h *ProviderHandler) changed() {
	if h.onChange != nil {
		h.onChange()
	}
}
