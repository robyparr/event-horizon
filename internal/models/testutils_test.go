package models

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	_ "github.com/lib/pq"
)

func newTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL_TEST"))
	if err != nil {
		t.Fatal(err)
	}

	migrationFiles, err := filepath.Glob("../../migrations/*.up.sql")
	if err != nil {
		t.Fatal(err)
	}

	for _, file := range migrationFiles {
		migration, err := os.ReadFile(file)
		if err != nil {
			db.Close()
			t.Fatal(err)
		}

		_, err = db.Exec(string(migration))
		if err != nil {
			db.Close()
			t.Fatal(err)
		}
	}

	t.Cleanup(func() {
		defer db.Close()

		migrationFiles, err := filepath.Glob("../../migrations/*.down.sql")
		if err != nil {
			db.Close()
			t.Fatal(err)
		}

		for _, file := range migrationFiles {
			migration, err := os.ReadFile(file)
			if err != nil {
				db.Close()
				t.Fatal(err)
			}

			_, err = db.Exec(string(migration))
			if err != nil {
				db.Close()
				t.Fatal(err)
			}
		}
	})

	return db
}
