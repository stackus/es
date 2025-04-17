package main

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/stackus/envelope"

	"github.com/stackus/es"
	"github.com/stackus/es/repositories/memory"
)

func main() {
	ctx := context.Background()

	// Create a new registry using the envelope package
	reg := envelope.NewRegistry()

	// Register the events and snapshot used by the Order aggregate
	if err := reg.Register(
		// Register the events used by the Order aggregate
		OrderCreated{},
		OrderItemAdded{},
		OrderItemRemoved{},
		OrderItemQuantityIncreased{},
		OrderItemQuantityDecreased{},
		OrderShippingAddressSet{},
		// Register the snapshot used by the Order aggregate
		OrderSnapshot{},
	); err != nil {
		panic(err)
	}

	// Create a new AggregateStore with a memory-based EventRepository and SnapshotRepository
	var store es.AggregateStore[uuid.UUID]
	store = es.NewSnapshotStore[uuid.UUID](
		reg,
		es.NewEventStore[uuid.UUID](
			reg,
			memory.NewEventRepository[uuid.UUID](),
		),
		memory.NewSnapshotRepository[uuid.UUID](),
		es.NewFrequencySnapshotStrategy[uuid.UUID](4), // Create a snapshot every 4 events
	)
	// Add a "logging" middleware to the store; we're simply printing to the console
	store = NewLoggingAggregateStore(store)

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
