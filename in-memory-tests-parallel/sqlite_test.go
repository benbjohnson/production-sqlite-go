package sqlite_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestInMemoryDB0(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB1(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB2(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB3(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB4(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB5(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB6(t *testing.T) { t.Parallel(); testInMemoryDB(t) }
func TestInMemoryDB7(t *testing.T) { t.Parallel(); testInMemoryDB(t) }

func testInMemoryDB(t *testing.T) {
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
