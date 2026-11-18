package hosts

import (
	"os"
	"path/filepath"
	"testing"

	"nextopsx/service/executor"
	"nextopsx/store"
	hoststore "nextopsx/store/hosts"
)

func openTestService(t *testing.T) *Service {
	t.Helper()
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	db, err := store.Open()
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.GORM().DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	hosts := hoststore.NewHostsStore(db)
	return New(hosts, executor.NewResolver(hosts), nil)
}

func TestHostCRUD(t *testing.T) {
	s := openTestService(t)
	folderID, err := s.CreateFolder("prod", "")
	if err != nil {
		t.Fatalf("CreateFolder: %v", err)
	}
	id, err := s.CreateHost(hoststore.HostInput{
		Name: "web-01", ParentID: folderID, Addr: "10.0.0.5", Port: 22,
		User: "root", AuthType: "password", Secret: "s3cret", Protocol: "ssh",
		AutoApprove: "inherit",
	})
	if err != nil {
		t.Fatalf("CreateHost: %v", err)
	}

	roots, err := s.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts: %v", err)
	}
	if len(roots) != 1 || roots[0].ID != folderID || len(roots[0].Children) != 1 {
		t.Fatalf("tree = %#v", roots)
	}
	if roots[0].Children[0].ID != id || roots[0].Children[0].Name != "web-01" {
		t.Fatalf("host = %#v", roots[0].Children[0])
	}

	secret, authType, err := s.GetHostSecret(id)
	if err != nil || secret != "s3cret" || authType != "password" {
		t.Fatalf("secret = %q/%q err=%v", secret, authType, err)
	}

	// Empty secret keeps the stored credential.
	if err := s.UpdateHost(id, hoststore.HostInput{
		Name: "web-02", ParentID: folderID, Addr: "10.0.0.6", Port: 2222,
		User: "ops", AuthType: "password", Protocol: "ssh",
	}); err != nil {
		t.Fatalf("UpdateHost: %v", err)
	}
	secret, _, err = s.GetHostSecret(id)
	if err != nil || secret != "s3cret" {
		t.Fatalf("secret after update = %q err=%v", secret, err)
	}

	if err := s.RenameNode(folderID, "production"); err != nil {
		t.Fatalf("RenameNode: %v", err)
	}
	if err := s.MoveNode(id, ""); err != nil {
		t.Fatalf("MoveNode: %v", err)
	}
	roots, _ = s.ListHosts()
	if len(roots) != 2 {
		t.Fatalf("root should hold folder and moved host: %#v", roots)
	}

	if err := s.DeleteNode(folderID); err != nil {
		t.Fatalf("DeleteNode: %v", err)
	}
	roots, _ = s.ListHosts()
	if len(roots) != 1 || roots[0].ID != id {
		t.Fatalf("tree after delete = %#v", roots)
	}
}

func TestTestConnectionSQLite(t *testing.T) {
	s := openTestService(t)
	path := filepath.Join(t.TempDir(), "t.db")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	f.Close()

	ok, err := s.TestConnection(hoststore.HostInput{
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
