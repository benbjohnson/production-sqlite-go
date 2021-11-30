package sqlite_test

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/tailscale/sqlite"
)

func TestInMemoryDB0(t *testing.T) {
	for i := 0; i < 1000; i++ {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			testInMemoryDB(t, i)
		})
	}
}

func testInMemoryDB(t *testing.T, i int) {
	t.Parallel()

	t0 := time.Now()
	defer func() { t.Logf("elapsed=%s", time.Since(t0)) }()

	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		t.Fatal(err)
	} else if _, err := db.Exec(`INSERT INTO t (name) VALUES ('jane')`); err != nil {
		t.Fatal(err)
	}

	var name string
	if err := db.QueryRow(`SELECT name FROM t WHERE id = 1`).Scan(&name); err != nil {
		t.Fatal(err)
	} else if got, want := name, "jane"; got != want {
		t.Fatalf("name=%q, want %q", got, want)
	}
}
