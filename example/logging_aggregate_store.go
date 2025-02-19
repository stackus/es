package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/stackus/es"
)

type LoggingAggregateStore struct {
	es.AggregateStore[uuid.UUID]
}

func NewLoggingAggregateStore(store es.AggregateStore[uuid.UUID]) *LoggingAggregateStore {
	return &LoggingAggregateStore{store}
}

func (s *LoggingAggregateStore) Load(ctx context.Context, aggregate es.Aggregate[uuid.UUID], hooks ...es.Hook[uuid.UUID]) error {

	hooks = append(hooks,
		// log after we've loaded events
		es.EventsPostLoadHook[uuid.UUID](func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], events []es.Event[uuid.UUID]) error {
			for _, event := range events {
				// "log" the event
				fmt.Println("Loaded event:", event.EventType, "Version:", event.AggregateVersion)
			}
			return nil
		}),
		// log after we've loaded a snapshot
		es.SnapshotPostLoadHook[uuid.UUID](func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], snapshot *es.Snapshot[uuid.UUID]) error {
			if snapshot != nil {
				// "log" the snapshot
				fmt.Println("Loaded snapshot:", snapshot.SnapshotType, "Version:", snapshot.AggregateVersion)
			}
			return nil
		}),
	)

	err := s.AggregateStore.Load(ctx, aggregate, hooks...)
	if err != nil {
		return err
	}

	fmt.Println("Loaded aggregate:", aggregate.AggregateType(), "ID:", aggregate.AggregateID(), "Version:", aggregate.AggregateVersion())

	return nil
}

func (s *LoggingAggregateStore) Save(ctx context.Context, aggregate es.Aggregate[uuid.UUID], hooks ...es.Hook[uuid.UUID]) error {

	hooks = append(hooks,
		// log before we save events
		es.EventsPreSaveHook[uuid.UUID](func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], events []es.Event[uuid.UUID]) error {
			for _, event := range events {
				// "log" the event
				fmt.Println("Saving event:", event.EventType, "Version:", event.AggregateVersion)
			}
			return nil
		}),
		// log before we save a snapshot
		es.SnapshotPreSaveHook[uuid.UUID](func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], snapshot es.Snapshot[uuid.UUID]) error {
			// "log" the snapshot
			fmt.Println("Saving snapshot:", snapshot.SnapshotType, "Version:", snapshot.AggregateVersion)
			return nil
		}),
	)

	fmt.Println("Saving aggregate:", aggregate.AggregateType(), "ID:", aggregate.AggregateID(), "Version:", aggregate.AggregateVersion())

	return s.AggregateStore.Save(ctx, aggregate, hooks...)
}
