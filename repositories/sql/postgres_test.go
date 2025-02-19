package sql_test

import (
	"context"
	"database/sql"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type postgresContainer struct {
	*postgres.PostgresContainer
	ConnectionString string
}

func createPostgresContainer(ctx context.Context) (*postgresContainer, error) {
	pgContainer, err := postgres.Run(ctx, "postgres:17-alpine",
		postgres.WithDatabase("eventsourcing"),
		postgres.WithUsername("postgres"),
		postgres.WithPassword("postgres"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		return nil, err
	}
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, err
	}

	return &postgresContainer{
		PostgresContainer: pgContainer,
		ConnectionString:  connStr,
	}, nil
}

func newPostgresDB(connStr string, files []string) (*sql.DB, error) {
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, err
	}

	// run the sequel in the schema/postgres_*.sql files
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

func usePostgres(ctx context.Context, files []string) (*sql.DB, func() error, error) {
	pgContainer, err := createPostgresContainer(ctx)
	if err != nil {
		return nil, nil, err
	}
	db, err := newPostgresDB(pgContainer.ConnectionString, files)
	if err != nil {
		return nil, nil, err
	}

	cleanup := func() error {
		if err := db.Close(); err != nil {
			return err
		}
		if err := pgContainer.Terminate(ctx); err != nil {
			return err
		}
		return nil
	}

	return db, cleanup, nil
}
