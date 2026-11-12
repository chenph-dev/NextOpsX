package backend

import (
	"os"
	"path/filepath"
	"testing"

	hoststore "nextopsx/internal/store/hosts"
)

func openTestBackend(t *testing.T) *Backend {
	t.Helper()
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	b, err := New()
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(b.Close)
	return b
}

func TestHostCRUD(t *testing.T) {
	b := openTestBackend(t)
	folderID, err := b.CreateFolder("prod", "")
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	id, err := b.CreateHost(hoststore.HostInput{
		Name: "web-01", ParentID: folderID, Addr: "10.0.0.5", Port: 22,
		User: "root", AuthType: "password", Secret: "s3cret", Protocol: "ssh",
		AutoApprove: "inherit",
	})
	if err != nil {
		t.Fatalf("CreateHost: %v", err)
	}

	roots, err := b.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts: %v", err)
	}
	if len(roots) != 1 || roots[0].ID != folderID || len(roots[0].Children) != 1 {
		t.Fatalf("tree = %#v", roots)
	}
	if roots[0].Children[0].ID != id || roots[0].Children[0].Name != "web-01" {
		t.Fatalf("host = %#v", roots[0].Children[0])
	}

	secret, authType, err := b.GetHostSecret(id)
	if err != nil || secret != "s3cret" || authType != "password" {
		t.Fatalf("secret = %q/%q err=%v", secret, authType, err)
	}

	// Empty secret keeps the stored credential.
	if err := b.UpdateHost(id, hoststore.HostInput{
		Name: "web-02", ParentID: folderID, Addr: "10.0.0.6", Port: 2222,
		User: "ops", AuthType: "password", Protocol: "ssh",
	}); err != nil {
		t.Fatalf("UpdateHost: %v", err)
	}
	secret, _, err = b.GetHostSecret(id)
	if err != nil || secret != "s3cret" {
		t.Fatalf("secret after update = %q err=%v", secret, err)
	}

	if err := b.RenameNode(folderID, "production"); err != nil {
		t.Fatalf("RenameNode: %v", err)
	}
	if err := b.MoveNode(id, ""); err != nil {
		t.Fatalf("MoveNode: %v", err)
	}
	roots, _ = b.ListHosts()
	if len(roots) != 2 {
		t.Fatalf("root should hold folder and moved host: %#v", roots)
	}

	if err := b.DeleteNode(folderID); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}
	roots, _ = b.ListHosts()
	if len(roots) != 1 || roots[0].ID != id {
		t.Fatalf("tree after delete = %#v", roots)
	}
}

func TestTestConnectionSQLite(t *testing.T) {
	b := openTestBackend(t)
	path := filepath.Join(t.TempDir(), "t.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	ok, err := b.TestConnection(hoststore.HostInput{
		Protocol: "sqlite",
		Params:   map[string]any{"filePath": path},
	})
	if err != nil {
		t.Fatalf("TestConnection: %v", err)
	}
	if !ok {
		t.Fatal("sqlite connection test should succeed")
	}
}
