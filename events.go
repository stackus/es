package es

import (
	"context"
	"time"

	"github.com/stackus/envelope"
)

type Event[K comparable] struct {
	AggregateID      K
	AggregateType    string
	AggregateVersion int
	EventType        string
	EventData        []byte
	OccurredAt       time.Time
}

type EventRepository[K comparable] interface {
	Load(ctx context.Context, aggregate AggregateRoot[K], hooks EventLoadHooks[K]) ([]Event[K], error)
	Save(ctx context.Context, aggregate AggregateRoot[K], events []Event[K], hooks EventSaveHooks[K]) error
}

type eventStore[K comparable] struct {
	registry   envelope.Registry
	repository EventRepository[K]
}

var _ AggregateStore[any] = (*eventStore[any])(nil)

func NewEventStore[K comparable](
	registry envelope.Registry,
	repository EventRepository[K],
) AggregateStore[K] {
	return &eventStore[K]{
		registry:   registry,
		repository: repository,
	}
}

func (s eventStore[K]) Load(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	events, err := s.repository.Load(ctx, aggregate, Hooks[K](hooks))
	if err != nil {
		return err
	}

	for _, event := range events {
		eEnv, err := s.registry.Deserialize(event.EventData)
		if err != nil {
			return err
		}
		err = aggregate.TrackChange(aggregate, eEnv.Payload())
		if err != nil {
			return err
		}
	}

	aggregate.commitChanges()

	return nil
}

func (s eventStore[K]) Save(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	changes := aggregate.Changes()

	if len(changes) == 0 {
		return nil
	}

	events := make([]Event[K], 0, len(changes))

	for i, change := range changes {
		eEnv, err := s.registry.Serialize(change)
		if err != nil {
			return err
		}
		events = append(events, Event[K]{
			AggregateID:      aggregate.AggregateID(),
			AggregateType:    aggregate.AggregateType(),
			AggregateVersion: aggregate.AggregateVersion() + i + 1,
			EventType:        eEnv.Key(),
			EventData:        eEnv.Bytes(),
			OccurredAt:       time.Now(),
		})
	}

	if err := s.repository.Save(ctx, aggregate, events, Hooks[K](hooks)); err != nil {
		return err
	}

	aggregate.commitChanges()

	return nil
}
