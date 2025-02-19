package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/stackus/es"
)

type snapshotRepository[K comparable] struct {
	db dbContext
}

func NewSnapshotRepository[K comparable](db *sql.DB) es.SnapshotRepository[K] {
	return &snapshotRepository[K]{
		db: dbContext{db},
	}
}

const loadSnapshotSQL = `SELECT aggregate_version, snapshot_type, snapshot_data, created_at
FROM aggregate_snapshots
WHERE aggregate_id = $1 AND aggregate_type = $2
ORDER BY aggregate_version DESC
LIMIT 1`

func (r *snapshotRepository[K]) Load(ctx context.Context, aggregate es.Aggregate[K], hooks es.SnapshotLoadHooks[K]) (*es.Snapshot[K], error) {
	var snapshot es.Snapshot[K]
	err := r.db.withTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// run preload hooks
		if err := hooks.SnapshotPreLoad(ctx, aggregate); err != nil {
			return err
		}

		// load snapshot
		row := tx.QueryRowContext(ctx, loadSnapshotSQL, aggregate.AggregateID(), aggregate.AggregateType())

		if err := row.Scan(&snapshot.AggregateVersion, &snapshot.SnapshotType, &snapshot.SnapshotData, &snapshot.CreatedAt); err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return err
			}
		}

		// run postload hooks
		if err := hooks.SnapshotPostLoad(ctx, aggregate, &snapshot); err != nil {
			return err
		}

		return nil
	})

	return &snapshot, err
}

const saveSnapshotSQL = `INSERT INTO aggregate_snapshots (aggregate_id, aggregate_type, aggregate_version, snapshot_type, snapshot_data, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`

func (r *snapshotRepository[K]) Save(ctx context.Context, aggregate es.Aggregate[K], snapshot es.Snapshot[K], hooks es.SnapshotSaveHooks[K]) error {
	return r.db.withTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// run presave hooks
		if err := hooks.SnapshotPreSave(ctx, aggregate, snapshot); err != nil {
			return err
		}

		// save snapshot
		_, err := tx.ExecContext(ctx, saveSnapshotSQL, snapshot.AggregateID, snapshot.AggregateType, snapshot.AggregateVersion, snapshot.SnapshotType, snapshot.SnapshotData, snapshot.CreatedAt)
		if err != nil {
			return err
		}

		// run postsave hooks
		if err := hooks.SnapshotPostSave(ctx, aggregate, snapshot); err != nil {
			return err
		}

		return nil
	})
}
