
# es &mdash; Event Sourcing in Go

[![Go Reference](https://pkg.go.dev/badge/github.com/stackus/es.svg)](https://pkg.go.dev/github.com/stackus/es)
[![Go Report Card](https://goreportcard.com/badge/github.com/stackus/es)](https://goreportcard.com/report/github.com/stackus/es)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

Event Sourcing library for Go, designed for building scalable, event-driven applications with **CQRS** and **DDD** principles.

## ✨ Features

- Lightweight Event Sourcing primitives for building event-driven systems.
- Supports generics for flexible aggregate IDs of nearly any Go type.
- Flexible Event Store abstractions for custom persistence layers.
- Built-in snapshot support to optimize aggregate state recovery.
- Pluggable aggregate root implementations for domain modeling.
- Support for multiple storage backends, including SQL and NoSQL.
- Extensible pre- and post-hook system for custom event processing.

## 🚀 Getting Started

### Installation

```sh
go get github.com/stackus/es
```

You will need to also install the required [envelope](https://github.com/stackus/envelope) package:

```sh
go get github.com/stackus/envelope
```

Envelope provides the serialization and deserialization of events and snapshots in a way that after they have been deserialized,
they can be used as the original type.

> TODO: Make it less of a manual process to use envelope. Possible: Replace the uses with an Optional parameter in the `NewEventStore` and `NewSnapshotStore` functions.

## 📚 Usage

There are a few basic concepts to understand when using this library:

- **Aggregate Root**: The primary entity in an event-sourced system.
- **Aggregate ID**: The unique identifier for an aggregate root.
- **Event**: A change that has occurred to an aggregate root.
- **Aggregate Store**: The persistence layer for storing and retrieving events and snapshots for an aggregate root.

### 1. Define an Aggregate

```go
// type Aggregate[K comparable] interface {
// 	AggregateType() string
// 	ApplyChange(event any) error
// }

type Order struct {
	es.AggregateRoot[uuid.UUID] // embed the AggregateRoot
	Total int
}

// implement the Aggregate[K] interface; implement the AggregateType method
func (o *Order) AggregateType() string { return "Order" }

// implement the ApplyChange method
func (o *Order) ApplyChange(event any) error {
	switch e := event.(type) {
	case *OrderCreated:
		o.Total = e.Total
	}
	return nil
}
```

### 2. Create an Aggregate ID

```go
// type AggregateID[K comparable] interface {
//	Get() K
//	New() K
//	Set(id K)
//	IsSet() bool
// }

type RootID uuid.UUID

// implement the AggregateID interface for the RootID type
func (r *RootID) Get() uuid.UUID    { return uuid.UUID(*r) }
func (r *RootID) New() uuid.UUID    { return uuid.New() }
func (r *RootID) Set(id uuid.UUID)  { *r = RootID(id) }
func (r *RootID) IsSet() bool       { return *r != RootID(uuid.Nil) }
```
You are free to use whatever kind of ID you want, as long as it you implement the `es.AggregateID[K]` interface.
There are tests and examples in this repository that show the usage of `string` and `int` IDs as well.

> TODO: Move docs for the ID before the Aggregate? It seems like it would be more logical to explain the ID before the Aggregate.

### 3. Define Events

```go
// a simple Go struct with exported fields will do just fine
type OrderCreated struct {
	Total   int
}
```

### 4. Create a Constructor or Factory Function

```go
// example of simple constructor
func NewOrder(id *RootID) *Order {
	return &Order{
		AggregateRoot: es.NewAggregateRoot(id),
	}
}

// example of factory function
func CreateOrder(id *RootID, total int) (*Order, error) {
	order := NewOrder(id)

	// record a change to the new aggregate
	if err := order.TrackChange(order, &OrderCreated{
		Total: total,
	}); err != nil {
		return nil, err
	}

	return order, nil
}
```

The `TrackChange(aggregate es.AggregateRoot[K], event any) error` method is used to apply changes to an aggregate root.
This method is provided by the embedded `es.AggregateRoot[K]` in the aggregate struct.

The changes are applied to the aggregate with the previously seen `ApplyChange(event any) error` method implemented in `Order`.

### 5. Create an Event Store

#### a. Create a repository that implements the `es.EventRepository` interface:

```go
repository := memory.NewEventRepository[uuid.UUID]()
```

#### b. All events that you want to store must be registered with a registry:

```go
reg := envelope.NewRegistry()
_ = reg.Register(OrderCreated{})
```

#### c. Create an instance of the event store:

```go
eventStore := es.NewEventStore(reg, repository)
```

### 6. Loading and Saving Events

To load all changes for an aggregate, you will do something similar to this:

```go
id := RootID(someOrderID)
order := NewOrder(&id)

err := eventStore.Load(ctx, order)
if err != nil {
	return err
}
```

To save uncommitted changes made to an aggregate, you will do something similar to this:

```go
err := eventStore.Save(ctx, order)
if err != nil {
	return err
}
```

Both of these methods will use the hooks you provide to process the events before and after they are saved or loaded.

Hooks are an optional third variadic parameter to the `Load` and `Save` methods.
The types of hooks available include pre-hooks and post-hooks, for example `EventsPreSave` and `EventsPostLoad`.

```go

var hooks []es.Hook[uuid.UUID]
// add a pre-save hook
hooks = append(hooks, es.EventsPreSaveHook(func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], events []es.Event[uuid.UUID]) error {
	// do something before saving
	return nil
}))
// add a post-save hook
hooks = append(hooks, es.EventsPostSaveHook(func(ctx context.Context, aggregate es.Aggregate[uuid.UUID], events []es.Event[uuid.UUID]) error {
	// do something after saving
	return nil
}))

err := eventStore.Save(ctx, order, hooks...)
```

Use these hooks to add custom behavior to the saving and loading of events. Logging, "domain events", and other behaviors can be added here.

## Snapshots

Snapshots are a way to optimize the loading of an aggregate by storing the state of the aggregate at a certain point in time.

Using Snapshots is entirely optional, but can be invaluable when you have aggregates with a large number of events.

### 1. Define a Snapshot

```go
type OrderSnapshot struct {
	Total int
}
```

### 2. Add the required methods to your Aggregate

```go
// type SnapshotAggregate[K comparable] interface {
// 	CreateSnapshot() any
//	ApplySnapshot(snapshot any) error
// }

func (o *Order) CreateSnapshot() any {
	return &OrderSnapshot{
		Total: o.Total,
	}
}

func (o *Order) ApplySnapshot(snapshot any) error {
	switch s := snapshot.(type) {
	case *OrderSnapshot:
		o.Total = s.Total
	}
	
	return nil
}
```

### 3. Create a Snapshot Store

#### a. Create a repository that implements the `es.SnapshotRepository` interface:

```go
snapshotRepository := memory.NewSnapshotRepository[uuid.UUID]()
```

#### b. Register the snapshot type(s) with the envelope registry:

```go
_ = reg.Register(OrderSnapshot{})
```

#### c. Create an instance of the snapshot store:

```go
snapshotStore := es.NewSnapshotStore(
	reg,
	snapshotRepository, 
	eventStore, // we will use the event store we created earlier to save events
	es.NewFrequencySnapshotStrategy(10), // create a new snapshot every 10 events
)
```

> There are other strategies available, such as `es.NewParticularChangesSnapshotStrategy(changes...)`, which creates a new snapshot when a particular change has occurred. Of course, you can also create your own strategy by implementing the `es.SnapshotStrategy` interface.

### 4. Loading and Saving Snapshots

The `SnapshotStore` has the same `Load` and `Save` methods as the `EventStore`. They both implement the `AggregateStore[K]` interface.

This means that we also have access to the same hooks that we used with the `EventStore`.
The only difference is that the hooks are applied to snapshots instead of events, so they are of type `SnapshotPre*` and `SnapshotPost*`.

These hooks are used by the snapshot store to hook into the saving and loading of events.

## 📜 License

This project is licensed under the MIT License—see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

Contributions, issues, and feature requests are welcome! Feel free to check the [issues page](https://github.com/stackus/es/issues).
