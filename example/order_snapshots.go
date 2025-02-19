package main

import (
	"github.com/google/uuid"
)

type OrderSnapshot struct {
	CustomerID uuid.UUID
	Items      []Item
	Shipping   ShippingAddress
}
