package backend

import (
	"archive/zip"
	"bytes"
	"strings"
	"testing"
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

func TestSkillLifecycle(t *testing.T) {
	b := openTestBackend(t)
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
