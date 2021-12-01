package main

import (
	"database/sql"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/sync/errgroup"
)

var (
	initialize = flag.Bool("init", false, "initialize schema & exit")
	n          = flag.Int("n", 1000, "iterations")
	parallel   = flag.Int("parallel", 1, "parallel goroutines")
)

const RowN = 100_000_000

const (
	ZipfS = 1.5
	ZipfV = 8
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	flag.Parse()
	if flag.NArg() == 0 {
		return fmt.Errorf("DSN required")
	} else if flag.NArg() > 1 {
		return fmt.Errorf("too many arguments")
	}
	dsn := flag.Arg(0)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(`PRAGMA busy_timeout = 5000;`); err != nil {
		return err
	}

	if *initialize {
		return initDB(db)
	}

	// Setup input.
	ch := make(chan struct{}, *n)
	for i := 0; i < *n; i++ {
		ch <- struct{}{}
	}
	close(ch)

	// Create distribution.
	zipf := rand.NewZipf(rand.New(rand.NewSource(0)), ZipfS, ZipfV, RowN)

	// Initialize connection.
	if err := runIter(db, 1); err != nil {
		return err
	}

	t0 := time.Now()

	// Execute in parallel.
	var g errgroup.Group
	for i := 0; i < *parallel; i++ {
		g.Go(func() error {
			for range ch {
				if err := runIter(db, int(zipf.Uint64())+1); err != nil {
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
	if _, err := db.Exec(`PRAGMA journal_mode = WAL;`); err != nil {
		return err
	}

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY, name TEXT);`); err != nil {
		return err
	}

	var index int
	for i := 0; i < 1000; i++ {
		if err := func() error {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			defer tx.Rollback()

			fmt.Printf("[inserting] %d\n", index)
			for j := 0; j < 100000; j++ {
				index++
				if _, err := tx.Exec(fmt.Sprintf(`INSERT INTO t (id, name) VALUES (%d, '%8x');`, index, index)); err != nil {
					return err
				}
			}

			return tx.Commit()
		}(); err != nil {
			return err
		}
	}
	fmt.Println("[done]")

	return nil
}

func runIter(db *sql.DB, id int) error {
	var name string
	if err := db.QueryRow(`SELECT name FROM t WHERE id = ?`, id).Scan(&name); err != nil {
		return err
	}
	return nil
}
