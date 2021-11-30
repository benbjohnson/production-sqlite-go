package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	_ "github.com/lib/pq"
	"golang.org/x/sync/errgroup"
	_ "modernc.org/sqlite"
)

var (
	driver = flag.String("driver", "", "driver name (postgres, sqlite)")
	dsn    = flag.String("dsn", "", "data source name,")

	initialize = flag.Bool("init", false, "initialize schema & exit")
	n          = flag.Int("n", 1000, "iterations")
	parallel   = flag.Int("parallel", 1, "parallel goroutines")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	flag.Parse()
	if *driver == "" {
		return fmt.Errorf("required: --driver")
	} else if *dsn == "" {
		return fmt.Errorf("required: --dsn")
	}

	db, err := sql.Open(*driver, *dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if *initialize {
		return initDB(db)
	}

	// Setup input.
	ch := make(chan struct{}, *n)
	for i := 0; i < *n; i++ {
		ch <- struct{}{}
	}
	close(ch)

	// Initialize connection.
	if err := runIter(db); err != nil {
		return err
	}

	t0 := time.Now()

	// Execute in parallel.
	var g errgroup.Group
	for i := 0; i < *parallel; i++ {
		g.Go(func() error {
			for range ch {
				if err := runIter(db); err != nil {
					return err
				}
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return err
	}

	elapsed := time.Since(t0)
	fmt.Printf("%d iterations completed in %fs (%s/iter)\n", *n, elapsed.Seconds(), (elapsed / time.Duration(*n)).String())
	return nil
}

func initDB(db *sql.DB) error {
	if *driver == "sqlite" {
		if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
			return err
		}
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`CREATE TABLE t (id INTEGER, name TEXT);`); err != nil {
		return err
	}

	for i := 0; i < 1000; i++ {
		if _, err := tx.Exec(fmt.Sprintf(`INSERT INTO t (id, name) VALUES (%d, '%8x');`, i, i)); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func runIter(db *sql.DB) error {
	var name string
	if err := db.QueryRow(`SELECT name FROM t WHERE id = 1`).Scan(&name); err != nil {
		return err
	}
	return nil
}
