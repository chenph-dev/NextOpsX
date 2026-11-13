package configstore

import (
	"testing"

	"nextopsx/internal/store"
	convstore "nextopsx/internal/store/conversations"
	hoststore "nextopsx/internal/store/hosts"
)

func TestProviderStore_CRUD_APIKeyEncrypted(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, _ := store.Open()
	defer closeDB(app)

	s := NewProviderStore(app)
	id, err := s.Create(AIProvider{
		Name: "公司 Claude", Protocol: "claude", Model: "claude-sonnet-5",
		BaseURL: "https://api.anthropic.com", APIKey: "sk-secret",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	var blob []byte
	app.GORM().Raw(`SELECT api_key_encrypted FROM ai_providers WHERE id=?`, id).Scan(&blob)
	if string(blob) == "sk-secret" {
		t.Fatal("API Key 明文存储")
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].APIKey != "" {
		t.Fatalf("列表不应返回 API Key: %+v", list)
	}
	if !list[0].IsDefault {
		t.Fatal("第一条供应商应自动成为默认")
	}

	got, err := s.Get(id)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.APIKey != "sk-secret" || got.Model != "claude-sonnet-5" {
		t.Fatalf("回读不匹配: %+v", got)
	}

	if err := s.Update(AIProvider{
		ID: id, Name: "公司 Claude 2", Protocol: "claude", Model: "claude-opus-5-5",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}
	got, _ = s.Get(id)
	if got.Name != "公司 Claude 2" || got.Model != "claude-opus-5-5" || got.APIKey != "" {
		t.Fatalf("更新后不匹配: %+v", got)
	}
}

func TestProviderStore_DefaultsAndDelete(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, _ := store.Open()
	defer closeDB(app)

	s := NewProviderStore(app)
	first, err := s.Create(AIProvider{Name: "a", Protocol: "openai", Model: "m1"})
	if err != nil {
		t.Fatalf("Create first: %v", err)
	}
	second, err := s.Create(AIProvider{Name: "b", Protocol: "openai", Model: "m2"})
	if err != nil {
		t.Fatalf("Create second: %v", err)
	}
	if _, err := s.Create(AIProvider{Name: "a", Protocol: "openai", Model: "dup"}); err == nil {
		t.Fatal("重名应报错")
	}

	def, err := s.Default()
	if err != nil || def.ID != first {
		t.Fatalf("默认应是第一条: %+v err=%v", def, err)
	}
	if err := s.SetDefault(second); err != nil {
		t.Fatalf("SetDefault: %v", err)
	}
	def, _ = s.Default()
	if def.ID != second {
		t.Fatalf("默认应切换为第二条: %+v", def)
	}

	// 删除默认供应商：剩下的自动成为默认。
	if err := s.Delete(second); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	def, _ = s.Default()
	if def.ID != first {
		t.Fatalf("删除默认后应回落到剩余供应商: %+v", def)
	}
	// 最后一条不允许删除。
	if err := s.Delete(first); err == nil {
		t.Fatal("应至少保留一个供应商")
	}
}

func TestProviderStore_DeleteClearsConversationRefs(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, _ := store.Open()
	defer closeDB(app)

	hostID, err := hoststore.NewHostsStore(app).SaveHost(hoststore.HostInput{
		Name: "h", Addr: "1.1.1.1", Port: 22, User: "u",
		AuthType: "password", Secret: "x",
	})
	if err != nil {
		t.Fatalf("SaveHost: %v", err)
	}
	s := NewProviderStore(app)
	used, err := s.Create(AIProvider{Name: "used", Protocol: "openai", Model: "m1"})
	if err != nil {
		t.Fatalf("Create used: %v", err)
	}
	if _, err := s.Create(AIProvider{Name: "spare", Protocol: "openai", Model: "m2"}); err != nil {
		t.Fatalf("Create spare: %v", err)
	}
	convs := convstore.NewConvStore(app)
	sid, err := convs.NewConversationWithProvider(hostID, "t", used)
	if err != nil {
		t.Fatalf("NewConversationWithProvider: %v", err)
	}

	if err := s.Delete(used); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	conv, err := convs.GetConversation(sid)
	if err != nil {
		t.Fatalf("GetConversation: %v", err)
	}
	if conv.ProviderID != "" {
		t.Fatalf("被删供应商的会话应回落默认，provider_id = %q", conv.ProviderID)
	}
}

func TestProviderStore_ResolveFallsBack(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, _ := store.Open()
	defer closeDB(app)

	s := NewProviderStore(app)
	if _, err := s.Resolve(""); err == nil {
		t.Fatal("没有供应商时应报错")
	}
	id, err := s.Create(AIProvider{Name: "a", Protocol: "openai", Model: "m1"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if p, err := s.Resolve(""); err != nil || p.ID != id {
		t.Fatalf("空引用应解析到默认: %+v err=%v", p, err)
	}
	if p, err := s.Resolve("missing"); err != nil || p.ID != id {
		t.Fatalf("已删除的引用应回落默认: %+v err=%v", p, err)
	}
}

func closeDB(app *store.DB) {
	sqlDB, _ := app.GORM().DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}
