package es

import (
	"context"
	"time"

	"github.com/stackus/envelope"
)

type (
	Snapshot[K comparable] struct {
		AggregateID      K
		AggregateType    string
		AggregateVersion int
		SnapshotType     string
		SnapshotData     []byte
		CreatedAt        time.Time
	}

	SnapshotAggregate[K comparable] interface {
		CreateSnapshot() any
		ApplySnapshot(snapshot any) error
		Aggregate[K]
	}

	SnapshotRepository[K comparable] interface {
		Load(ctx context.Context, aggregate AggregateRoot[K], hooks SnapshotLoadHooks[K]) (*Snapshot[K], error)
		Save(ctx context.Context, aggregate AggregateRoot[K], snapshot Snapshot[K], hooks SnapshotSaveHooks[K]) error
	}

	snapshotStore[K comparable] struct {
		registry   envelope.Registry
		eventStore AggregateStore[K]
		repository SnapshotRepository[K]
		strategy   SnapshotStrategy[K]
	}
)

var _ AggregateStore[any] = (*snapshotStore[any])(nil)

func NewSnapshotStore[K comparable](
	registry envelope.Registry,
	eventStore AggregateStore[K],
	repository SnapshotRepository[K],
	strategy SnapshotStrategy[K],
) AggregateStore[K] {
	return &snapshotStore[K]{
		registry:   registry,
		eventStore: eventStore,
		repository: repository,
		strategy:   strategy,
	}
}

func (s snapshotStore[K]) Load(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	hook := EventsPreLoadHook[K](func(ctx context.Context, aggregate AggregateRoot[K]) error {
		sa, ok := aggregate.(SnapshotAggregate[K])
		if !ok {
			return nil
		}

		snapshot, err := s.repository.Load(ctx, aggregate, Hooks[K](hooks))
		if err != nil || snapshot == nil {
			return err
		}

		ssEnv, err := s.registry.Deserialize(snapshot.SnapshotData)
		if err != nil {
			return err
		}
		err = sa.ApplySnapshot(ssEnv.Payload())
		if err != nil {
			return err
		}
		sa.SetID(snapshot.AggregateID)
		sa.SetVersion(snapshot.AggregateVersion)

		return nil
	})

	return s.eventStore.Load(ctx, aggregate, append([]Hook[K]{hook}, hooks...)...)
}

func (s snapshotStore[K]) Save(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	hook := EventsPostSaveHook[K](func(ctx context.Context, aggregate AggregateRoot[K], events []Event[K]) error {
		sa, ok := aggregate.(SnapshotAggregate[K])
		if !ok {
			return nil
		}

		if !s.strategy.ShouldSnapshot(aggregate) {
			return nil
		}

		snapshot := sa.CreateSnapshot()
		ssEnv, err := s.registry.Serialize(snapshot)
		if err != nil {
			return err
		}

		return s.repository.Save(ctx, aggregate, Snapshot[K]{
			AggregateID:      aggregate.AggregateID(),
			AggregateType:    aggregate.AggregateType(),
			AggregateVersion: aggregate.AggregateVersion() + len(events),
			SnapshotType:     ssEnv.Key(),
			SnapshotData:     ssEnv.Bytes(),
			CreatedAt:        time.Now(),
		}, Hooks[K](hooks))
	})

	return s.eventStore.Save(ctx, aggregate, append(hooks, hook)...)
}
