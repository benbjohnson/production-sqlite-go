package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/felixge/fgprof"
	// _ "github.com/mattn/go-sqlite3"
	_ "github.com/tailscale/sqlite"
	"golang.org/x/sync/errgroup"
)

var (
	n          = flag.Int("n", 1000, "iterations")
	ops        = flag.Int("ops", 1, "operations per iteration")
	parallel   = flag.Int("parallel", 1, "parallel goroutines")
	fgprofPath = flag.String("fgprof", "", "fgprof output")
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

func run() error {
	flag.Parse()

	// Setup input.
	ch := make(chan struct{}, *n)
	for i := 0; i < *n; i++ {
		ch <- struct{}{}
	}
	close(ch)

	// Run profiling, if specified.
	if *fgprofPath != "" {
		f, err := os.Create(*fgprofPath)
		if err != nil {
			return err
		}
		defer f.Close()

		defer fgprof.Start(f, fgprof.FormatPprof)()
	}

	t0 := time.Now()

	// Execute in parallel.
	var g errgroup.Group
	for i := 0; i < *parallel; i++ {
		g.Go(func() error {
			for range ch {
				if err := runIter(); err != nil {
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

func runIter() error {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}
	defer db.Close()

	if _, err := db.Exec(`CREATE TABLE t (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT)`); err != nil {
		return err
	} else if _, err := db.Exec(`INSERT INTO t (name) VALUES ('jane')`); err != nil {
		return err
	}

	for i := 0; i < *ops; i++ {
		var name string
		if err := db.QueryRow(`SELECT name FROM t WHERE id = 1`).Scan(&name); err != nil {
			return err
		}
	}

	return nil
}
