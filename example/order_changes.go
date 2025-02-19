package main

import (
	"github.com/google/uuid"
)

type OrderCreated struct {
	CustomerID uuid.UUID
}

type OrderItemAdded struct {
	ProductID uuid.UUID
	Quantity  int
}

type OrderItemRemoved struct {
	ProductID uuid.UUID
}

type OrderItemQuantityIncreased struct {
	ProductID uuid.UUID
	Delta     int
}

type OrderItemQuantityDecreased struct {
	ProductID uuid.UUID
	Delta     int
}

type OrderShippingAddressSet struct {
	Street     string
	City       string
	State      string
	PostalCode string
}
