package es

import (
	"reflect"
)

type (
	SnapshotStrategy[K comparable] interface {
		ShouldSnapshot(aggregate Aggregate[K]) bool
	}

	frequencySnapshotStrategy[K comparable] struct {
		frequency int
	}

	particularChangesSnapshotStrategy[K comparable] struct {
		changes []any
	}
)

var _ SnapshotStrategy[any] = (*frequencySnapshotStrategy[any])(nil)
var _ SnapshotStrategy[any] = (*particularChangesSnapshotStrategy[any])(nil)

// NewFrequencySnapshotStrategy creates a new SnapshotStrategy that takes in a frequency value.
//
// This strategy is useful when you want to snapshot the aggregate every N changes.
func NewFrequencySnapshotStrategy[K comparable](frequency int) SnapshotStrategy[K] {
	return &frequencySnapshotStrategy[K]{
		frequency: frequency,
	}
}

func (ss frequencySnapshotStrategy[K]) ShouldSnapshot(aggregate Aggregate[K]) bool {
	return (aggregate.AggregateVersion()%ss.frequency)+len(aggregate.Changes()) >= ss.frequency
}

// NewParticularChangesSnapshotStrategy creates a new SnapshotStrategy that takes in a list of changes.
//
// This strategy is useful when you want to snapshot the aggregate when a particular change is applied.
func NewParticularChangesSnapshotStrategy[K comparable](changes []any) SnapshotStrategy[K] {
	return &particularChangesSnapshotStrategy[K]{
		changes: changes,
	}
}

func (ss particularChangesSnapshotStrategy[K]) ShouldSnapshot(aggregate Aggregate[K]) bool {
	for _, change := range ss.changes {
		t1 := reflect.TypeOf(change)
		if t1.Kind() == reflect.Ptr {
			t1 = t1.Elem()
		}

		for _, aggregateChange := range aggregate.Changes() {
			t2 := reflect.TypeOf(aggregateChange)
			if t2.Kind() == reflect.Ptr {
				t2 = t2.Elem()
			}
			if t1 == t2 {
				return true
			}
		}
	}

	return false
}
