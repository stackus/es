package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/stackus/es"
	"github.com/stackus/es/repositories/memory"
)

func main() {
	ctx := context.Background()

	// Create a new AggregateStore with a memory-based EventRepository and SnapshotRepository
	var store es.AggregateStore[uuid.UUID]
	eventStore := es.NewEventStore[uuid.UUID](memory.NewEventRepository[uuid.UUID]())

	// Register the events with the event store directly (won't use reflection)
	es.RegisterEvent(eventStore, &OrderCreated{})
	es.RegisterEvent(eventStore, &OrderItemAdded{})

	// Register the events with the event store using a slice of EventPayloads (will use reflection)
	events := []es.EventPayload{
		&OrderItemRemoved{}, // in a slice it won't matter if you use a pointer or value type
		&OrderItemQuantityIncreased{},
		OrderItemQuantityDecreased{},
		OrderShippingAddressSet{},
	}

	for _, event := range events {
		es.RegisterEvent(eventStore, event)
	}

	snapshotStore := es.NewSnapshotStore[uuid.UUID](
		eventStore,
		memory.NewSnapshotRepository[uuid.UUID](),
		es.NewFrequencySnapshotStrategy[uuid.UUID](4), // Create a snapshot every 4 events
	)
	// Register the snapshot with the snapshot store
	es.RegisterSnapshot(snapshotStore, &OrderSnapshot{})

	// Add a "logging" middleware to the store; we're simply printing to the console
	store = NewLoggingAggregateStore(snapshotStore)

	// Create a new Order aggregate (creates the aggregate and applies the OrderCreated event)
	customerID := uuid.New()
	order, _ := CreateOrder(customerID)
	orderID := order.AggregateID() // Save the order ID for later

	// Add an item to the order (applies the OrderItemAdded event)
	productID := uuid.New()
	if err := order.AddItem(productID, 1); err != nil {
		panic(err)
	}

	// Set the shipping address on the order (applies the OrderShippingAddressSet event)
	if err := order.SetShippingAddress("123 Main St", "City", "ZZ", "00123"); err != nil {
		panic(err)
	}

	// Save the order to the store
	if err := store.Save(ctx, order); err != nil {
		panic(err)
	}

	// Load the order from the store
	order = NewOrder(orderID)
	if err := store.Load(ctx, order); err != nil {
		panic(err)
	}

	// Change the quantity of the item in the order (applies the OrderItemQuantityIncreased event)
	if err := order.ChangeItemQuantity(productID, 2); err != nil {
		panic(err)
	}

	// Save the order to the store (this time a snapshot will be created)
	if err := store.Save(ctx, order); err != nil {
		panic(err)
	}

	// Load the order from the store; this time we will load from the snapshot
	order = NewOrder(orderID)
	if err := store.Load(ctx, order); err != nil {
		panic(err)
	}

	// Add a second item to the order (applies the OrderItemAdded event)
	productID = uuid.New()
	if err := order.AddItem(productID, 1); err != nil {
		panic(err)
	}
	// Change the quantity of the second item in the order (applies the OrderItemQuantityIncreased event)
	if err := order.ChangeItemQuantity(productID, 2); err != nil {
		panic(err)
	}

	// Save the order to the store
	if err := store.Save(ctx, order); err != nil {
		panic(err)
	}

	// Load the order from the store once more; this time we will load from the snapshot and events
	order = NewOrder(orderID)
	if err := store.Load(ctx, order); err != nil {
		panic(err)
	}

	// Print the order details
	fmt.Printf("\n\nOrderID: %s\n", order.AggregateID())
	fmt.Printf("CustomerID: %s\n", order.CustomerID)
	fmt.Printf("Shipping Address: %s, %s, %s %s\n", order.ShippingAddress.Street, order.ShippingAddress.City, order.ShippingAddress.State, order.ShippingAddress.PostalCode)
	for _, item := range order.Items {
		fmt.Printf("ProductID: %s, Quantity: %d\n", item.ProductID, item.Quantity)
	}
}
