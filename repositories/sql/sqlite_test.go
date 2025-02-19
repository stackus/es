package sql_test

import (
	"context"
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

func newSQLiteDB(files []string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", "file:memdb1?mode=memory&cache=shared")
	if err != nil {
		return nil, err
	}

	// run the sequel in the schema/sqlite_*.sql files
	for _, file := range files {
		content, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		_, err = db.Exec(string(content))
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

func useSqlite(ctx context.Context, files []string) (*sql.DB, func() error, error) {
	db, err := newSQLiteDB(files)
	if err != nil {
		return nil, nil, err
	}

	return db, db.Close, nil
}
