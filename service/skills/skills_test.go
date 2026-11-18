package skills

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"nextopsx/domain/skill"
	"nextopsx/store"
	skillsstore "nextopsx/store/skills"
)

func testSkillZip(t *testing.T) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	write := func(name, content string) {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	write("disk-analysis/SKILL.md", "---\nname: disk-analysis\ndescription: 分析磁盘占用\n---\n# 磁盘分析\n\n运行 df -h。\n")
	write("disk-analysis/scripts/check.sh", "#!/bin/sh\necho ok\n")
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

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
	root := filepath.Join(t.TempDir(), "skills")
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	mgr := skill.NewManager(skillsstore.NewSkillStore(data), root)
	return New(mgr, nil)
}

func TestSkillLifecycle(t *testing.T) {
	b := openTestService(t)
	s, err := b.InstallSkill(testSkillZip(t))
	if err != nil {
		t.Fatalf("InstallSkill: %v", err)
	}
	if s.Name != "disk-analysis" || !s.Enabled {
		t.Fatalf("skill = %+v", s)
	}
	items, err := b.ListSkills()
	if err != nil || len(items) != 1 || items[0].Name != "disk-analysis" {
		t.Fatalf("list = %+v err=%v", items, err)
	}
	md, err := b.ReadSkillMarkdown("disk-analysis")
	if err != nil || !strings.Contains(md, "# 磁盘分析") {
		t.Fatalf("markdown = %q err=%v", md, err)
	}
	if err := b.SetSkillEnabled("disk-analysis", false); err != nil {
		t.Fatalf("SetSkillEnabled: %v", err)
	}
	items, _ = b.ListSkills()
	if len(items) != 1 || items[0].Enabled {
		t.Fatalf("enabled state = %+v", items)
	}
	if err := b.DeleteSkill("disk-analysis"); err != nil {
		t.Fatalf("DeleteSkill: %v", err)
	}
	items, _ = b.ListSkills()
	if len(items) != 0 {
		t.Fatalf("after delete = %+v", items)
	}
}
