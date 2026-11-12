package backend

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	hoststore "nextopsx/internal/store/hosts"
)

func TestDatabaseSchemaAndQuerySQLite(t *testing.T) {
	b := openTestBackend(t)
	path := filepath.Join(t.TempDir(), "test.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`CREATE TABLE users (id INTEGER PRIMARY KEY, name TEXT, note TEXT)`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO users (name, note) VALUES ('alice', 'first'), ('bob', NULL)`); err != nil {
		t.Fatal(err)
	}
	db.Close()

	hostID, err := b.CreateHost(hoststore.HostInput{
		Name: "db", Protocol: "sqlite",
		Params: map[string]any{"filePath": path},
	})
	if err != nil {
		t.Fatal(err)
	}

	schema, err := b.ListSchema(hostID)
	if err != nil {
		t.Fatalf("ListSchema: %v", err)
	}
	if len(schema.Tables) != 1 || schema.Tables[0].Name != "users" {
		t.Fatalf("schema = %+v", schema)
	}
	if len(schema.Tables[0].Columns) != 3 {
		t.Fatalf("columns = %+v", schema.Tables[0].Columns)
	}

	res, err := b.ExecuteSQL(hostID, `SELECT name, note FROM users ORDER BY id`)
	if err != nil {
		t.Fatalf("ExecuteSQL: %v", err)
	}
	if len(res.Columns) != 2 || len(res.Rows) != 2 {
		t.Fatalf("result = %+v", res)
	}
	if res.Rows[0][0] != "alice" || res.Rows[1][0] != "bob" {
		t.Fatalf("rows = %+v", res.Rows)
	}
	if res.Rows[1][1] != nil {
		t.Fatalf("NULL cell = %#v", res.Rows[1][1])
	}

	execRes, err := b.ExecuteSQL(hostID, `INSERT INTO users (name) VALUES ('carol')`)
	if err != nil {
		t.Fatalf("ExecuteSQL insert: %v", err)
	}
	if execRes.RowsAffected != 1 {
		t.Fatalf("rowsAffected = %d", execRes.RowsAffected)
	}
}
