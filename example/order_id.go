package main

import (
	"github.com/google/uuid"

	"github.com/stackus/es"
)

type OrderID uuid.UUID

func (id *OrderID) Get() uuid.UUID {
	return uuid.UUID(*id)
}

func (id *OrderID) New() uuid.UUID {
	return uuid.New()
}

func (id *OrderID) IsSet() bool {
	return *id != OrderID(uuid.Nil)
}

func (id *OrderID) Set(newID uuid.UUID) {
	if *id != OrderID(uuid.Nil) {
		return
	}
	*id = OrderID(newID)
}

var _ es.AggregateID[uuid.UUID] = (*OrderID)(nil)
