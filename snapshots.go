package es

import (
	"context"
	"encoding/json"
	"time"
)

type (
	SnapshotPayload interface {
		Kind() string
	}

	Snapshot[K comparable] struct {
		AggregateID      K
		AggregateType    string
		AggregateVersion int
		SnapshotType     string
		SnapshotData     []byte
		CreatedAt        time.Time
	}

	SnapshotAggregate[K comparable] interface {
		CreateSnapshot() SnapshotPayload
		ApplySnapshot(snapshot SnapshotPayload) error
		Aggregate[K]
	}

	SnapshotRepository[K comparable] interface {
		Load(ctx context.Context, aggregate AggregateRoot[K], hooks SnapshotLoadHooks[K]) (*Snapshot[K], error)
		Save(ctx context.Context, aggregate AggregateRoot[K], snapshot Snapshot[K], hooks SnapshotSaveHooks[K]) error
	}

	SnapshotStore[K comparable] struct {
		eventStore       AggregateStore[K]
		repository       SnapshotRepository[K]
		strategy         SnapshotStrategy[K]
		snapshotAppliers map[string]snapshotApplier[K]
	}
)

var _ AggregateStore[any] = (*SnapshotStore[any])(nil)

func NewSnapshotStore[K comparable](
	eventStore *EventStore[K],
	repository SnapshotRepository[K],
	strategy SnapshotStrategy[K],
) *SnapshotStore[K] {
	return &SnapshotStore[K]{
		eventStore:       eventStore,
		repository:       repository,
		strategy:         strategy,
		snapshotAppliers: make(map[string]snapshotApplier[K]),
	}
}

func (s *SnapshotStore[K]) Load(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	hook := EventsPreLoadHook[K](func(ctx context.Context, aggregate AggregateRoot[K]) error {
		sa, ok := aggregate.(SnapshotAggregate[K])
		if !ok {
			return nil
		}

		snapshot, err := s.repository.Load(ctx, aggregate, Hooks[K](hooks))
		if err != nil || snapshot == nil {
			return err
		}

		if applier, ok := s.snapshotAppliers[snapshot.SnapshotType]; ok {
			if err := applier.applyChange(snapshot, sa); err != nil {
				return err
			}
		} else {
			return ErrUnregisteredSnapshot(snapshot.SnapshotType)
		}
		sa.SetID(snapshot.AggregateID)
		sa.SetVersion(snapshot.AggregateVersion)

		return nil
	})

	return s.eventStore.Load(ctx, aggregate, append([]Hook[K]{hook}, hooks...)...)
}

func (s *SnapshotStore[K]) Save(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	hook := EventsPostSaveHook[K](func(ctx context.Context, aggregate AggregateRoot[K], events []Event[K]) error {
		sa, ok := aggregate.(SnapshotAggregate[K])
		if !ok {
			return nil
		}

		if !s.strategy.ShouldSnapshot(aggregate) {
			return nil
		}

		snapshot := sa.CreateSnapshot()
		data, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}

		return s.repository.Save(ctx, aggregate, Snapshot[K]{
			AggregateID:      aggregate.AggregateID(),
			AggregateType:    aggregate.AggregateType(),
			AggregateVersion: aggregate.AggregateVersion() + len(events),
			SnapshotType:     snapshot.Kind(),
			SnapshotData:     data,
			CreatedAt:        time.Now(),
		}, Hooks[K](hooks))
	})

	return s.eventStore.Save(ctx, aggregate, append(hooks, hook)...)
}

func (s *SnapshotStore[K]) registerSnapshotApplier(snapshotType string, applier snapshotApplier[K]) {
	s.snapshotAppliers[snapshotType] = applier
}

type snapshotApplier[K comparable] interface {
	applyChange(snapshot *Snapshot[K], aggregate SnapshotAggregate[K]) error
}

type snapshotApplyWrapper[K comparable, T SnapshotPayload] struct{}

func (a *snapshotApplyWrapper[K, T]) applyChange(snapshot *Snapshot[K], aggregate SnapshotAggregate[K]) error {
	var payload T
	if err := json.Unmarshal(snapshot.SnapshotData, &payload); err != nil {
		return err
	}
	return aggregate.ApplySnapshot(payload)
}

func RegisterSnapshot[K comparable, T SnapshotPayload](snapshotStore *SnapshotStore[K], snapshot T) {
	snapshotStore.registerSnapshotApplier(snapshot.Kind(), &snapshotApplyWrapper[K, T]{})
}
