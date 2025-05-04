package es

import (
	"context"
)

type (
	AggregateID[K comparable] interface {
		Get() K      // Get the current ID; this can return a zero value and IsSet should be used to check for zero values
		New() K      // Generate new ID based on type-specific rules
		Set(K)       // Set the ID if it is not already set
		IsSet() bool // Check against type-specific zero value rules
	}

	AggregateRoot[K comparable] interface {
		Aggregate[K]
		AggregateType() string
		ApplyChange(event EventPayload) error
	}

	AggregateStore[K comparable] interface {
		Load(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error
		Save(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error
	}

	Aggregate[K comparable] interface {
		AggregateID() K
		AggregateVersion() int
		TrackChange(aggregate AggregateRoot[K], event EventPayload) error
		Changes() []EventPayload
		SetID(K)
		SetVersion(int)
		commitChanges()
	}

	aggregate[K comparable] struct {
		id      AggregateID[K]
		version int
		changes []EventPayload
	}
)

func NewAggregate[K comparable](id AggregateID[K]) Aggregate[K] {
	return &aggregate[K]{
		id:      id,
		version: 0,
		changes: []EventPayload{},
	}
}

func (a *aggregate[K]) AggregateID() K          { return a.id.Get() }
func (a *aggregate[K]) AggregateVersion() int   { return a.version }
func (a *aggregate[K]) Changes() []EventPayload { return a.changes }
func (a *aggregate[K]) TrackChange(aggregate AggregateRoot[K], event EventPayload) error {
	if !a.id.IsSet() {
		a.SetID(a.id.New())
	}

	a.changes = append(a.changes, event)
	return aggregate.ApplyChange(event)
}

func (a *aggregate[K]) SetID(id K) {
	if !a.id.IsSet() {
		a.id.Set(id)
	}
}

func (a *aggregate[K]) SetVersion(version int) { a.version = version }

func (a *aggregate[K]) commitChanges() {
	a.version += len(a.changes)
	a.changes = make([]EventPayload, 0)
}
