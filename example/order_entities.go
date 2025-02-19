package main

import (
	"github.com/google/uuid"
)

type Item struct {
	ProductID uuid.UUID
	Quantity  int
}

type ShippingAddress struct {
	Street     string
	City       string
	State      string
	PostalCode string
}
