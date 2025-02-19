package memory

import (
	"context"

	"github.com/stackus/es"
)

type eventRepository[K comparable] struct {
	streams map[string]map[K][]es.Event[K]
}

func NewEventRepository[K comparable]() es.EventRepository[K] {
	return &eventRepository[K]{
		streams: make(map[string]map[K][]es.Event[K]),
	}
}

func (r *eventRepository[K]) Load(ctx context.Context, aggregate es.Aggregate[K], hooks es.EventLoadHooks[K]) ([]es.Event[K], error) {
	if err := hooks.EventsPreLoad(ctx, aggregate); err != nil {
		return nil, err
	}

	stream := map[K][]es.Event[K]{
		aggregate.AggregateID(): make([]es.Event[K], 0),
	}
	if s, ok := r.streams[aggregate.AggregateType()]; ok {
		stream = s
	}

	var events []es.Event[K]

	if e, ok := stream[aggregate.AggregateID()]; ok {
		events = e
	}

	fromVersion := aggregate.AggregateVersion()
	// return a partial list from index fromVersion to the end
	switch {
	case fromVersion >= len(events):
		events = nil
	case fromVersion > 0:
		events = events[fromVersion:]
	}
	// if fromVersion > 0 {
	// 	if fromVersion >= len(events) {
	// 		return nil, nil
	// 	}
	// 	events = events[fromVersion:]
	// }

	if err := hooks.EventsPostLoad(ctx, aggregate, events); err != nil {
		return nil, err
	}

	return events, nil
}

func (r *eventRepository[K]) Save(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K], hooks es.EventSaveHooks[K]) error {
	if err := hooks.EventsPreSave(ctx, aggregate, events); err != nil {
		return err
	}

	for _, event := range events {
		stream, ok := r.streams[event.AggregateType]
		if !ok {
			stream = make(map[K][]es.Event[K])
			r.streams[event.AggregateType] = stream
		}

		// ensure that the version of the event lines up with its position in the stream
		if len(stream[event.AggregateID])+1 != event.AggregateVersion {
			return es.ErrAggregateVersionConflict{
				AggregateID:      event.AggregateID,
				AggregateType:    event.AggregateType,
				AggregateVersion: event.AggregateVersion,
			}
		}
		stream[event.AggregateID] = append(stream[event.AggregateID], event)
	}

	return hooks.EventsPostSave(ctx, aggregate, events)
}
