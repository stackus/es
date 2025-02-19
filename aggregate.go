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

	Aggregate[K comparable] interface {
		AggregateRoot[K]
		AggregateType() string
		ApplyChange(event any) error
	}

	AggregateStore[K comparable] interface {
		Load(ctx context.Context, aggregate Aggregate[K], hooks ...Hook[K]) error
		Save(ctx context.Context, aggregate Aggregate[K], hooks ...Hook[K]) error
	}

	AggregateRoot[K comparable] interface {
		Root() AggregateRoot[K]
		AggregateID() K
		AggregateVersion() int
		TrackChange(aggregate Aggregate[K], event any) error
		Changes() []any
		SetID(K)
		commitChanges()
		setVersion(int)
	}

	aggregateRoot[K comparable] struct {
		id      AggregateID[K]
		version int
		changes []any
	}
)

func NewAggregateRoot[K comparable](id AggregateID[K]) AggregateRoot[K] {
	return &aggregateRoot[K]{
		id:      id,
		version: 0,
		changes: []any{},
	}
}

func (a *aggregateRoot[K]) Root() AggregateRoot[K] { return a }
func (a *aggregateRoot[K]) AggregateID() K         { return a.id.Get() }
func (a *aggregateRoot[K]) AggregateVersion() int  { return a.version }
func (a *aggregateRoot[K]) Changes() []any         { return a.changes }
func (a *aggregateRoot[K]) TrackChange(aggregate Aggregate[K], event any) error {
	if !a.id.IsSet() {
		a.SetID(a.id.New())
	}

	a.changes = append(a.changes, event)
	return aggregate.ApplyChange(event)
}

func (a *aggregateRoot[K]) SetID(id K) {
	if !a.id.IsSet() {
		a.id.Set(id)
	}
}

func (a *aggregateRoot[K]) commitChanges() {
	a.version += len(a.changes)
	a.changes = make([]any, 0)
}

func (a *aggregateRoot[K]) setVersion(version int) { a.version = version }
