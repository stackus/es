package es

import (
	"context"
	"encoding/json"
	"reflect"
	"time"
)

type (
	EventPayload interface {
		Kind() string
	}

	Event[K comparable] struct {
		AggregateID      K
		AggregateType    string
		AggregateVersion int
		EventType        string
		EventData        []byte
		OccurredAt       time.Time
	}

	EventRepository[K comparable] interface {
		Load(ctx context.Context, aggregate AggregateRoot[K], hooks EventLoadHooks[K]) ([]Event[K], error)
		Save(ctx context.Context, aggregate AggregateRoot[K], events []Event[K], hooks EventSaveHooks[K]) error
	}

	EventStore[K comparable] struct {
		repository    EventRepository[K]
		eventAppliers map[string]eventApplier[K]
		eventHooks    []Hook[K]
	}
)

var _ AggregateStore[any] = (*EventStore[any])(nil)

func NewEventStore[K comparable](
	repository EventRepository[K],
	hooks ...Hook[K],
) *EventStore[K] {
	return &EventStore[K]{
		repository:    repository,
		eventAppliers: make(map[string]eventApplier[K]),
		eventHooks:    hooks,
	}
}

func (s *EventStore[K]) Load(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	events, err := s.repository.Load(ctx, aggregate, Hooks[K](append(hooks, s.eventHooks...)))
	if err != nil {
		return err
	}

	for _, event := range events {
		if applier, ok := s.eventAppliers[event.EventType]; ok {
			err = applier.applyChange(event, aggregate)
			if err != nil {
				return err
			}
			continue
		} else {
			return ErrUnregisteredEvent(event.EventType)
		}
	}

	aggregate.commitChanges()

	return nil
}

func (s *EventStore[K]) Save(ctx context.Context, aggregate AggregateRoot[K], hooks ...Hook[K]) error {
	changes := aggregate.Changes()

	if len(changes) == 0 {
		return nil
	}

	events := make([]Event[K], 0, len(changes))

	for i, change := range changes {
		data, err := json.Marshal(change)
		if err != nil {
			return err
		}
		events = append(events, Event[K]{
			AggregateID:      aggregate.AggregateID(),
			AggregateType:    aggregate.AggregateType(),
			AggregateVersion: aggregate.AggregateVersion() + i + 1,
			EventType:        change.Kind(),
			EventData:        data,
			OccurredAt:       time.Now(),
		})
	}

	if err := s.repository.Save(ctx, aggregate, events, Hooks[K](append(hooks, s.eventHooks...))); err != nil {
		return err
	}

	aggregate.commitChanges()

	return nil
}

// WithRepository returns a new EventStore with the provided repository.
func (s *EventStore[K]) WithRepository(repository EventRepository[K]) *EventStore[K] {
	return &EventStore[K]{
		repository:    repository,
		eventAppliers: s.eventAppliers,
		eventHooks:    s.eventHooks,
	}
}

func (s *EventStore[K]) registerEventApplier(eventType string, applier eventApplier[K]) {
	if s.eventAppliers == nil {
		s.eventAppliers = make(map[string]eventApplier[K])
	}
	s.eventAppliers[eventType] = applier
}

type eventApplier[K comparable] interface {
	applyChange(event Event[K], aggregate AggregateRoot[K]) error
}

type eventTypedWrapper[K comparable, T EventPayload] struct{}

func (a *eventTypedWrapper[K, T]) applyChange(event Event[K], aggregate AggregateRoot[K]) error {
	var payload T
	if err := json.Unmarshal(event.EventData, &payload); err != nil {
		return err
	}
	return aggregate.TrackChange(aggregate, payload)
}

type eventReflectWrapper[K comparable] struct {
	v reflect.Value
}

func (w *eventReflectWrapper[K]) applyChange(event Event[K], aggregate AggregateRoot[K]) error {
	payload := w.v.Interface().(EventPayload)
	if err := json.Unmarshal(event.EventData, &payload); err != nil {
		return err
	}
	return aggregate.TrackChange(aggregate, payload)
}

func RegisterEvent[K comparable, T EventPayload](eventStore *EventStore[K], event T) {
	if reflect.TypeOf((*T)(nil)).Elem().Kind() == reflect.Interface {
		eventType := reflect.TypeOf(event)
		if eventType.Kind() == reflect.Ptr {
			eventType = eventType.Elem()
		}
		eventStore.registerEventApplier(event.Kind(), &eventReflectWrapper[K]{reflect.New(eventType)})
	} else {
		eventStore.registerEventApplier(event.Kind(), &eventTypedWrapper[K, T]{})
		return
	}
}
