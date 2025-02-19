package memory

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/stackus/es"
)

type hookMock[K comparable] struct {
	mock.Mock
}

var _ es.Hook[any] = (*hookMock[any])(nil)

func (m *hookMock[K]) EventsPreSave(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K]) error {
	args := m.Called(ctx, aggregate, events)
	return args.Error(0)
}

func (m *hookMock[K]) EventsPostSave(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K]) error {
	args := m.Called(ctx, aggregate, events)
	return args.Error(0)
}

func (m *hookMock[K]) EventsPreLoad(ctx context.Context, aggregate es.Aggregate[K]) error {
	args := m.Called(ctx, aggregate)
	return args.Error(0)
}

func (m *hookMock[K]) EventsPostLoad(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K]) error {
	args := m.Called(ctx, aggregate, events)
	return args.Error(0)
}

func (m *hookMock[K]) SnapshotPreSave(ctx context.Context, aggregate es.Aggregate[K], snapshot es.Snapshot[K]) error {
	args := m.Called(ctx, aggregate, snapshot)
	return args.Error(0)
}

func (m *hookMock[K]) SnapshotPostSave(ctx context.Context, aggregate es.Aggregate[K], snapshot es.Snapshot[K]) error {
	args := m.Called(ctx, aggregate, snapshot)
	return args.Error(0)
}

func (m *hookMock[K]) SnapshotPreLoad(ctx context.Context, aggregate es.Aggregate[K]) error {
	args := m.Called(ctx, aggregate)
	return args.Error(0)
}

func (m *hookMock[K]) SnapshotPostLoad(ctx context.Context, aggregate es.Aggregate[K], snapshot *es.Snapshot[K]) error {
	args := m.Called(ctx, aggregate, snapshot)
	return args.Error(0)
}
