package snippets

import (
	"testing"

	"nextopsx/store"
	snippetsstore "nextopsx/store/snippets"
)

func TestSnippetFacade(t *testing.T) {
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
	b := New(snippetsstore.NewSnippetStore(data))
	created, err := b.SaveCommandSnippet(snippetsstore.Snippet{
		Name: "磁盘", Group: "排障", Command: "df -h",
	})
	if err != nil {
		t.Fatalf("SaveCommandSnippet: %v", err)
	}
	if created.ID == "" {
		t.Fatal("SaveCommandSnippet should assign an ID")
	}
	list, err := b.ListCommandSnippets()
	if err != nil {
		t.Fatalf("ListCommandSnippets: %v", err)
	}
	if len(list) != 1 || list[0].Name != "磁盘" {
		t.Fatalf("list = %+v", list)
	}
	if err := b.DeleteCommandSnippet(created.ID); err != nil {
		t.Fatalf("DeleteCommandSnippet: %v", err)
	}
	list, err = b.ListCommandSnippets()
	if err != nil {
		t.Fatalf("ListCommandSnippets after delete: %v", err)
	}
	if len(list) != 0 {
		t.Fatalf("list after delete = %+v", list)
	}
}
