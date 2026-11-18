package policy

import (
	"testing"

	"nextopsx/store"
	configstore "nextopsx/store/config"
)

func openTestService(t *testing.T) *Service {
	t.Helper()
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	data, err := store.Open()
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := data.GORM().DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	return New(configstore.NewPolicyStore(data), nil)
}

func TestApprovalPolicyFacade(t *testing.T) {
	s := openTestService(t)

	p, err := s.GetApprovalPolicy()
	if err != nil {
		t.Fatalf("GetApprovalPolicy: %v", err)
	}
	if !p.EnableAuto {
		t.Fatalf("默认应开启自动放行: %+v", p)
	}

	if err := s.SaveApprovalPolicy(configstore.ApprovalPolicy{
		EnableAuto:   false,
		ReadOnlyList: []string{"df -h", "free -m"},
	}); err != nil {
		t.Fatalf("SaveApprovalPolicy: %v", err)
	}
	got, err := s.GetApprovalPolicy()
	if err != nil {
		t.Fatalf("GetApprovalPolicy: %v", err)
	}
	if got.EnableAuto || len(got.ReadOnlyList) != 2 {
		t.Fatalf("保存后不匹配: %+v", got)
	}
}
