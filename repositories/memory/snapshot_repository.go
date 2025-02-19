package memory

import (
	"context"

	"github.com/stackus/es"
)

type snapshotRepository[K comparable] struct {
	snapshots map[string]map[K]es.Snapshot[K]
}

func NewSnapshotRepository[K comparable]() es.SnapshotRepository[K] {
	return &snapshotRepository[K]{
		snapshots: make(map[string]map[K]es.Snapshot[K]),
	}
}

func (r *snapshotRepository[K]) Load(ctx context.Context, aggregate es.Aggregate[K], hooks es.SnapshotLoadHooks[K]) (*es.Snapshot[K], error) {
	if err := hooks.SnapshotPreLoad(ctx, aggregate); err != nil {
		return nil, err
	}

	stream, ok := r.snapshots[aggregate.AggregateType()]
	if !ok {
		return nil, nil
	}

	snapshot, ok := stream[aggregate.AggregateID()]
	if !ok {
		return nil, nil
	}

	if err := hooks.SnapshotPostLoad(ctx, aggregate, &snapshot); err != nil {
		return nil, err
	}

	return &snapshot, nil
}

func (r *snapshotRepository[K]) Save(ctx context.Context, aggregate es.Aggregate[K], snapshot es.Snapshot[K], hooks es.SnapshotSaveHooks[K]) error {
	if err := hooks.SnapshotPreSave(ctx, aggregate, snapshot); err != nil {
		return err
	}

	stream, ok := r.snapshots[snapshot.AggregateType]
	if !ok {
		stream = make(map[K]es.Snapshot[K])
		r.snapshots[snapshot.AggregateType] = stream
	}

	stream[snapshot.AggregateID] = snapshot

	return hooks.SnapshotPostSave(ctx, aggregate, snapshot)
}
