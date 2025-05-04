package main

import (
	"github.com/google/uuid"
)

type OrderCreated struct {
	CustomerID uuid.UUID
}

func (OrderCreated) Kind() string { return "OrderCreated" }

type OrderItemAdded struct {
	ProductID uuid.UUID
	Quantity  int
}

func (OrderItemAdded) Kind() string { return "OrderItemAdded" }

type OrderItemRemoved struct {
	ProductID uuid.UUID
}

func (OrderItemRemoved) Kind() string { return "OrderItemRemoved" }

type OrderItemQuantityIncreased struct {
	ProductID uuid.UUID
	Delta     int
}

func (OrderItemQuantityIncreased) Kind() string { return "OrderItemQuantityIncreased" }

type OrderItemQuantityDecreased struct {
	ProductID uuid.UUID
	Delta     int
}

func (OrderItemQuantityDecreased) Kind() string { return "OrderItemQuantityDecreased" }

type OrderShippingAddressSet struct {
	Street     string
	City       string
	State      string
	PostalCode string
}

func (OrderShippingAddressSet) Kind() string { return "OrderShippingAddressSet" }
