package snippetsstore

import (
	"testing"

	"nextopsx/internal/store"
)

func newTestStore(t *testing.T) *SnippetStore {
	t.Helper()
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	app, err := store.Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() {
		if sqlDB, _ := app.GORM().DB(); sqlDB != nil {
			sqlDB.Close()
		}
	})
	return NewSnippetStore(app)
}

func TestSnippetStoreCRUD(t *testing.T) {
	s := newTestStore(t)
	created, err := s.Save(Snippet{
		Name: "磁盘占用", Group: "排障", Command: "df -h",
		Description: "查看磁盘使用情况",
	})
	if err != nil {
		t.Fatalf("Save create: %v", err)
	}
	if created.ID == "" || created.CreatedAt == 0 {
		t.Fatalf("created = %+v", created)
	}

	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(list) != 1 || list[0].Name != "磁盘占用" || list[0].Group != "排障" {
		t.Fatalf("list = %+v", list)
	}

	created.Name = "磁盘使用"
	created.Command = "df -hT"
	updated, err := s.Save(created)
	if err != nil {
		t.Fatalf("Save update: %v", err)
	}
	if updated.Name != "磁盘使用" || updated.Command != "df -hT" {
		t.Fatalf("updated = %+v", updated)
	}
	if updated.CreatedAt != created.CreatedAt {
		t.Fatalf("created_at changed: %d -> %d", created.CreatedAt, updated.CreatedAt)
	}

	if err := s.Delete(created.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	list, err = s.List()
	if err != nil {
		t.Fatalf("List after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("list after delete = %+v", list)
	}
}

func TestSnippetStoreOrdersByGroupAndName(t *testing.T) {
	s := newTestStore(t)
	for _, in := range []Snippet{
		{Name: "b", Group: "z", Command: "echo b"},
		{Name: "a", Group: "z", Command: "echo a"},
		{Name: "c", Group: "a", Command: "echo c"},
	} {
		if _, err := s.Save(in); err != nil {
			t.Fatalf("Save %q: %v", in.Name, err)
		}
	}
	list, err := s.List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	want := []string{"c", "a", "b"}
	for i, name := range want {
		if list[i].Name != name {
			t.Fatalf("list[%d] = %q, want %q: %+v", i, list[i].Name, name, list)
		}
	}
}

func TestSnippetStoreValidation(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.Save(Snippet{Name: " ", Command: "ls"}); err == nil {
		t.Fatal("empty name should fail")
	}
	if _, err := s.Save(Snippet{Name: "ls", Command: " "}); err == nil {
		t.Fatal("empty command should fail")
	}
	if err := s.Delete("missing"); err == nil {
		t.Fatal("deleting a missing snippet should fail")
	}
	if _, err := s.Save(Snippet{ID: "missing", Name: "x", Command: "ls"}); err == nil {
		t.Fatal("updating a missing snippet should fail")
	}
}
