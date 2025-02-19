package main

import (
	"github.com/google/uuid"

	"github.com/stackus/es"
)

var _ es.Aggregate[uuid.UUID] = (*Order)(nil)
var _ es.SnapshotAggregate[uuid.UUID] = (*Order)(nil)

type Order struct {
	es.AggregateRoot[uuid.UUID]
	CustomerID      uuid.UUID
	Items           map[uuid.UUID]*Item
	ShippingAddress ShippingAddress
}

func NewOrder() *Order {
	order := &Order{
		AggregateRoot: es.NewAggregateRoot(&OrderID{}),
		Items:         make(map[uuid.UUID]*Item),
	}

	return order
}

// --- Domain methods ---

func CreateOrder(customerID uuid.UUID) (*Order, error) {
	order := NewOrder()

	return order, order.TrackChange(order, &OrderCreated{
		CustomerID: customerID,
	})
}

func (o *Order) AddItem(productID uuid.UUID, quantity int) error {
	return o.TrackChange(o, &OrderItemAdded{
		ProductID: productID,
		Quantity:  quantity,
	})
}

func (o *Order) RemoveItem(productID uuid.UUID) error {
	return o.TrackChange(o, &OrderItemRemoved{
		ProductID: productID,
	})
}

func (o *Order) ChangeItemQuantity(productID uuid.UUID, quantity int) error {
	if _, ok := o.Items[productID]; !ok {
		return nil
	}

	delta := quantity - o.Items[productID].Quantity
	if delta == 0 {
		return nil
	}
	if delta > 0 {
		return o.TrackChange(o, &OrderItemQuantityIncreased{
			ProductID: productID,
			Delta:     delta,
		})
	}

	return o.TrackChange(o, &OrderItemQuantityDecreased{
		ProductID: productID,
		Delta:     -delta,
	})
}

func (o *Order) SetShippingAddress(street, city, state, postalCode string) error {
	return o.TrackChange(o, &OrderShippingAddressSet{
		Street:     street,
		City:       city,
		State:      state,
		PostalCode: postalCode,
	})
}

// --- Event Sourcing methods ---

func (o *Order) AggregateType() string {
	return "orders.Order"
}

func (o *Order) ApplyChange(event any) error {
	switch e := event.(type) {
	case *OrderCreated:
		o.CustomerID = e.CustomerID
	case *OrderItemAdded:
		if _, ok := o.Items[e.ProductID]; !ok {
			o.Items[e.ProductID] = &Item{
				ProductID: e.ProductID,
				Quantity:  e.Quantity,
			}
		} else {
			o.Items[e.ProductID].Quantity += e.Quantity
		}
	case *OrderItemRemoved:
		delete(o.Items, e.ProductID)
	case *OrderItemQuantityIncreased:
		o.Items[e.ProductID].Quantity += e.Delta
	case *OrderItemQuantityDecreased:
		o.Items[e.ProductID].Quantity -= e.Delta
	case *OrderShippingAddressSet:
		o.ShippingAddress = ShippingAddress{
			Street:     e.Street,
			City:       e.City,
			State:      e.State,
			PostalCode: e.PostalCode,
		}
	default:
		return es.ErrUnknownAggregateChangeType{
			AggregateType: o.AggregateType(),
			Change:        event,
		}
	}

	return nil
}

func (o *Order) CreateSnapshot() any {
	items := make([]Item, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, *item)
	}
	return &OrderSnapshot{
		CustomerID: o.CustomerID,
		Items:      items,
		Shipping:   o.ShippingAddress,
	}
}

func (o *Order) ApplySnapshot(snapshot any) error {
	switch s := snapshot.(type) {
	case *OrderSnapshot:
		o.CustomerID = s.CustomerID
		o.Items = make(map[uuid.UUID]*Item, len(s.Items))
		for _, item := range s.Items {
			o.Items[item.ProductID] = &Item{
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
		}
		o.ShippingAddress = s.Shipping
	default:
		return es.ErrUnknownAggregateSnapshotType{
			AggregateType: o.AggregateType(),
			Snapshot:      snapshot,
		}
	}

	return nil
}
