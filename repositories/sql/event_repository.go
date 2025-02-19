package sql

import (
	"context"
	"database/sql"
	"errors"

	"github.com/stackus/es"
)

type eventRepository[K comparable] struct {
	db dbContext
}

var _ es.EventRepository[any] = (*eventRepository[any])(nil)

func NewEventRepository[K comparable](db *sql.DB) es.EventRepository[K] {
	return &eventRepository[K]{
		db: dbContext{db},
	}
}

const loadEventsSQL = `SELECT aggregate_version, event_type, event_data, occurred_at
FROM aggregate_events
WHERE aggregate_id = $1 AND aggregate_type = $2 AND aggregate_version > $3
ORDER BY aggregate_version ASC`

func (r *eventRepository[K]) Load(ctx context.Context, aggregate es.Aggregate[K], hooks es.EventLoadHooks[K]) ([]es.Event[K], error) {
	var events []es.Event[K]
	err := r.db.withTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// run preload hooks
		if err := hooks.EventsPreLoad(ctx, aggregate); err != nil {
			return err
		}

		// load events
		rows, err := tx.QueryContext(ctx, loadEventsSQL, aggregate.AggregateID(), aggregate.AggregateType(), aggregate.AggregateVersion())
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var event es.Event[K]
			event.AggregateID = aggregate.AggregateID()
			event.AggregateType = aggregate.AggregateType()
			if err := rows.Scan(&event.AggregateVersion, &event.EventType, &event.EventData, &event.OccurredAt); err != nil {
				return err
			}
			events = append(events, event)
		}

		// run postload hooks
		if err := hooks.EventsPostLoad(ctx, aggregate, events); err != nil {
			return err
		}

		return nil
	})

	return events, err
}

const saveEventsSQL = `INSERT INTO aggregate_events (aggregate_id, aggregate_type, aggregate_version, event_type, event_data, occurred_at)
VALUES ($1, $2, $3, $4, $5, $6)`

func (r *eventRepository[K]) Save(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K], hooks es.EventSaveHooks[K]) error {
	return r.db.withTx(ctx, func(ctx context.Context, tx *sql.Tx) error {
		// run presave hooks
		if err := hooks.EventsPreSave(ctx, aggregate, events); err != nil {
			return err
		}

		// save events
		for _, event := range events {
			if _, err := tx.ExecContext(ctx, saveEventsSQL, aggregate.AggregateID(), aggregate.AggregateType(), event.AggregateVersion, event.EventType, event.EventData, event.OccurredAt); err != nil {
				return err
			}
		}

		// run postsave hooks
		if err := hooks.EventsPostSave(ctx, aggregate, events); err != nil {
			return err
		}

		return nil
	})
}
