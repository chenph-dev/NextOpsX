package ai

import configstore "nextopsx/store/config"

// providerHandler 处理 AI 供应商的 CRUD；任何变更都会通知会话管理器失效模型缓存。
type providerHandler struct {
	providers *configstore.ProviderStore
	onChange  func()
}

// newProviderHandler 构造 providerHandler。onChange 可为 nil。
func newProviderHandler(providers *configstore.ProviderStore, onChange func()) *providerHandler {
	return &providerHandler{providers: providers, onChange: onChange}
}

// List 返回全部供应商（不含 API Key）。
func (h *providerHandler) List() ([]configstore.AIProvider, error) {
	return h.providers.List()
}

// Get 返回单个供应商（含 API Key，供编辑表单使用）。
func (h *providerHandler) Get(id string) (configstore.AIProvider, error) {
	return h.providers.Get(id)
}

// Create 新增供应商，返回新 ID。
func (h *providerHandler) Create(p configstore.AIProvider) (string, error) {
	id, err := h.providers.Create(p)
	if err != nil {
		return "", err
	}
	h.changed()
	return id, nil
}

// Update 更新供应商。
func (h *providerHandler) Update(p configstore.AIProvider) error {
	if err := h.providers.Update(p); err != nil {
		return err
	}
	h.changed()
	return nil
}

// Delete 删除供应商；引用它的历史对话回落默认供应商。
func (h *providerHandler) Delete(id string) error {
	if err := h.providers.Delete(id); err != nil {
		return err
	}
	h.changed()
	return nil
}

// SetDefault 把 id 设为默认供应商。
func (h *providerHandler) SetDefault(id string) error {
	if err := h.providers.SetDefault(id); err != nil {
		return err
	}
	h.changed()
	return nil
}

func (h *providerHandler) changed() {
	if h.onChange != nil {
		h.onChange()
	}
}
