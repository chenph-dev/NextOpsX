package resources

import "testing"

func TestIconsLoad(t *testing.T) {
	if FolderSVG() == nil {
		t.Fatal("folder icon missing")
	}
	for _, protocol := range []string{
		"mysql", "postgres", "sqlite", "sqlserver", "clickhouse",
		"redis", "elasticsearch", "ssh", "winrm",
	} {
		if ProtocolBitmap(protocol) == nil {
			t.Errorf("icon missing for %q", protocol)
		}
	}
	if ProtocolBitmap("ssh") != ProtocolBitmap("linux") {
		t.Error("ssh should use the Linux icon")
	}
	if ProtocolBitmap("winrm") != ProtocolBitmap("windows") {
		t.Error("winrm should use the Windows icon")
	}
	if ProtocolBitmap("unknown-protocol") != nil {
		t.Error("unknown protocol should have no icon")
	}
	for _, name := range []string{"table", "view", "field"} {
		if ObjectSVG(name) == nil {
			t.Errorf("object icon missing for %q", name)
		}
	}
}
