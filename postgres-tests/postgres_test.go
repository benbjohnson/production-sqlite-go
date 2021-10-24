package postgres_test

import (
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
)

func TestInMemoryDB(t *testing.T) {
	t0 := time.Now()

	db, err := sql.Open("postgres", "postgres://benbjohnson:@localhost/testdb?sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	t1 := time.Now()

	if _, err := db.Exec(`DROP TABLE IF EXISTS t`); err != nil {
		t.Fatal(err)
	}

	t2 := time.Now()

	if _, err := db.Exec(`CREATE TABLE t (id SERIAL PRIMARY KEY, name TEXT)`); err != nil {
		t.Fatal(err)
	}

	t3 := time.Now()

	if _, err := db.Exec(`INSERT INTO t (name) VALUES ('jane')`); err != nil {
		t.Fatal(err)
	}

	t4 := time.Now()

	var name string
	if err := db.QueryRow(`SELECT name FROM t WHERE id = 1`).Scan(&name); err != nil {
		t.Fatal(err)
	} else if got, want := name, "jane"; got != want {
		t.Fatalf("name=%q, want %q", got, want)
	}

	t5 := time.Now()

	t.Logf(" ELAPSED TIME")
	t.Logf("connect %s", t1.Sub(t0))
	t.Logf("drop    %s", t2.Sub(t1))
	t.Logf("create  %s", t3.Sub(t2))
	t.Logf("insert  %s", t4.Sub(t3))
	t.Logf("select  %s", t5.Sub(t4))
	t.Logf("TOTAL   %s", time.Since(t0))
}
