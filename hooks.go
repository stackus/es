package es

import (
	"context"
)

type (
	EventsPreSave[K comparable]  func(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error
	EventsPostSave[K comparable] func(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error

	EventsPreLoad[K comparable]  func(ctx context.Context, aggregate Aggregate[K]) error
	EventsPostLoad[K comparable] func(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error

	SnapshotPreSave[K comparable]  func(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error
	SnapshotPostSave[K comparable] func(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error

	SnapshotPreLoad[K comparable]  func(ctx context.Context, aggregate Aggregate[K]) error
	SnapshotPostLoad[K comparable] func(ctx context.Context, aggregate Aggregate[K], snapshot *Snapshot[K]) error

	EventSaveHooks[K comparable] interface {
		EventsPreSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error
		EventsPostSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error
	}

	EventLoadHooks[K comparable] interface {
		EventsPreLoad(ctx context.Context, aggregate Aggregate[K]) error
		EventsPostLoad(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error
	}

	SnapshotSaveHooks[K comparable] interface {
		SnapshotPreSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error
		SnapshotPostSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error
	}

	SnapshotLoadHooks[K comparable] interface {
		SnapshotPreLoad(ctx context.Context, aggregate Aggregate[K]) error
		SnapshotPostLoad(ctx context.Context, aggregate Aggregate[K], snapshot *Snapshot[K]) error
	}

	Hook[K comparable] interface {
		EventSaveHooks[K]
		EventLoadHooks[K]
		SnapshotSaveHooks[K]
		SnapshotLoadHooks[K]
	}

	hook[K comparable] struct {
		eventsPreSave    EventsPreSave[K]
		eventsPostSave   EventsPostSave[K]
		eventsPreLoad    EventsPreLoad[K]
		eventsPostLoad   EventsPostLoad[K]
		snapshotPreSave  SnapshotPreSave[K]
		snapshotPostSave SnapshotPostSave[K]
		snapshotPreLoad  SnapshotPreLoad[K]
		snapshotPostLoad SnapshotPostLoad[K]
	}

	Hooks[K comparable] []Hook[K]

	// hooks[K comparable] struct {
	// 	hooks []Hook[K]
	// }
)

func EventsPreSaveHook[K comparable](fn EventsPreSave[K]) Hook[K] {
	return &hook[K]{eventsPreSave: fn}
}

func (h *hook[K]) EventsPreSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	if h.eventsPreSave != nil {
		return h.eventsPreSave(ctx, aggregate, events)
	}
	return nil
}

func EventsPostSaveHook[K comparable](fn EventsPostSave[K]) Hook[K] {
	return &hook[K]{eventsPostSave: fn}
}

func (h *hook[K]) EventsPostSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	if h.eventsPostSave != nil {
		return h.eventsPostSave(ctx, aggregate, events)
	}
	return nil
}

func EventsPreLoadHook[K comparable](fn EventsPreLoad[K]) Hook[K] {
	return &hook[K]{eventsPreLoad: fn}
}

func (h *hook[K]) EventsPreLoad(ctx context.Context, aggregate Aggregate[K]) error {
	if h.eventsPreLoad != nil {
		return h.eventsPreLoad(ctx, aggregate)
	}
	return nil
}

func EventsPostLoadHook[K comparable](fn EventsPostLoad[K]) Hook[K] {
	return &hook[K]{eventsPostLoad: fn}
}

func (h *hook[K]) EventsPostLoad(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	if h.eventsPostLoad != nil {
		return h.eventsPostLoad(ctx, aggregate, events)
	}
	return nil
}

func SnapshotPreSaveHook[K comparable](fn SnapshotPreSave[K]) Hook[K] {
	return &hook[K]{snapshotPreSave: fn}
}

func (h *hook[K]) SnapshotPreSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error {
	if h.snapshotPreSave != nil {
		return h.snapshotPreSave(ctx, aggregate, snapshot)
	}
	return nil
}

func SnapshotPostSaveHook[K comparable](fn SnapshotPostSave[K]) Hook[K] {
	return &hook[K]{snapshotPostSave: fn}
}

func (h *hook[K]) SnapshotPostSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error {
	if h.snapshotPostSave != nil {
		return h.snapshotPostSave(ctx, aggregate, snapshot)
	}
	return nil
}

func SnapshotPreLoadHook[K comparable](fn SnapshotPreLoad[K]) Hook[K] {
	return &hook[K]{snapshotPreLoad: fn}
}

func (h *hook[K]) SnapshotPreLoad(ctx context.Context, aggregate Aggregate[K]) error {
	if h.snapshotPreLoad != nil {
		return h.snapshotPreLoad(ctx, aggregate)
	}
	return nil
}

func SnapshotPostLoadHook[K comparable](fn SnapshotPostLoad[K]) Hook[K] {
	return &hook[K]{snapshotPostLoad: fn}
}

func (h *hook[K]) SnapshotPostLoad(ctx context.Context, aggregate Aggregate[K], snapshot *Snapshot[K]) error {
	if h.snapshotPostLoad != nil {
		return h.snapshotPostLoad(ctx, aggregate, snapshot)
	}
	return nil
}

func (h Hooks[K]) EventsPreSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	for _, hook := range h {
		if err := hook.EventsPreSave(ctx, aggregate, events); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) EventsPostSave(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	for _, hook := range h {
		if err := hook.EventsPostSave(ctx, aggregate, events); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) EventsPreLoad(ctx context.Context, aggregate Aggregate[K]) error {
	for _, hook := range h {
		if err := hook.EventsPreLoad(ctx, aggregate); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) EventsPostLoad(ctx context.Context, aggregate Aggregate[K], events []Event[K]) error {
	for _, hook := range h {
		if err := hook.EventsPostLoad(ctx, aggregate, events); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) SnapshotPreSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error {
	for _, hook := range h {
		if err := hook.SnapshotPreSave(ctx, aggregate, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) SnapshotPostSave(ctx context.Context, aggregate Aggregate[K], snapshot Snapshot[K]) error {
	for _, hook := range h {
		if err := hook.SnapshotPostSave(ctx, aggregate, snapshot); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) SnapshotPreLoad(ctx context.Context, aggregate Aggregate[K]) error {
	for _, hook := range h {
		if err := hook.SnapshotPreLoad(ctx, aggregate); err != nil {
			return err
		}
	}
	return nil
}

func (h Hooks[K]) SnapshotPostLoad(ctx context.Context, aggregate Aggregate[K], snapshot *Snapshot[K]) error {
	for _, hook := range h {
		if err := hook.SnapshotPostLoad(ctx, aggregate, snapshot); err != nil {
			return err
		}
	}
	return nil
}
