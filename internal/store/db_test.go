package store

import "testing"

func TestOpen_CreatesAndMigrates(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer closeDB(s)

	var name string
	err = s.GORM().Raw(`SELECT name FROM sqlite_master WHERE type='table' AND name='commands_fts'`).Scan(&name).Error
	if err != nil {
		t.Fatalf("查询 FTS 表失败: %v", err)
	}
	if name != "commands_fts" {
		t.Fatalf("期望 commands_fts，得到 %s", name)
	}
}

func closeDB(s *DB) {
	sqlDB, _ := s.GORM().DB()
	if sqlDB != nil {
		sqlDB.Close()
	}
}

func TestOpen_SQLitePragmas(t *testing.T) {
	t.Setenv("NEXTOPSX_DATA_DIR", t.TempDir())
	t.Setenv("NEXTOPSX_TEST_NO_KEYRING", "1")
	s, err := Open()
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer closeDB(s)

	var mode string
	if err := s.GORM().Raw(`PRAGMA journal_mode`).Scan(&mode).Error; err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("journal_mode = %q, want wal", mode)
	}
	var timeout int
	if err := s.GORM().Raw(`PRAGMA busy_timeout`).Scan(&timeout).Error; err != nil {
		t.Fatal(err)
	}
	if timeout != 5000 {
		t.Fatalf("busy_timeout = %d, want 5000", timeout)
	}
	var fk int
	if err := s.GORM().Raw(`PRAGMA foreign_keys`).Scan(&fk).Error; err != nil {
		t.Fatal(err)
	}
	if fk != 1 {
		t.Fatalf("foreign_keys = %d, want 1", fk)
	}
}
