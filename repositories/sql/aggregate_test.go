package sql_test

import (
	"time"

	"github.com/google/uuid"
	"github.com/stackus/envelope"

	"github.com/stackus/es"
)

type (
	Record[K comparable] struct {
		es.Aggregate[K]
		Text      string
		Number    int
		Timestamp time.Time
	}

	RecordUUID struct {
		id uuid.UUID
	}

	RecordInt struct {
		id int
	}

	RecordString struct {
		id string
	}

	RecordCreated struct {
		Text      string
		Number    int
		Timestamp time.Time
	}

	RecordTextUpdated struct {
		Text string
	}

	RecordNumberUpdated struct {
		Number int
	}

	RecordTimestampUpdated struct {
		Timestamp time.Time
	}

	RecordSnapshot struct {
		Text      string
		Number    int
		Timestamp time.Time
	}
)

func (o *RecordUUID) New() uuid.UUID {
	return uuid.New()
}
func (o *RecordUUID) Get() uuid.UUID {
	return o.id
}
func (o *RecordUUID) IsSet() bool {
	return o.id != uuid.Nil
}
func (o *RecordUUID) Set(id uuid.UUID) {
	if o.id != uuid.Nil {
		return
	}
	o.id = id
}

var seq int = 0

func (o *RecordInt) New() int {
	seq++
	return seq
}
func (o *RecordInt) Get() int {
	return o.id
}
func (o *RecordInt) IsSet() bool {
	return o.id != 0
}
func (o *RecordInt) Set(id int) {
	if o.id != 0 {
		return
	}
	o.id = id
}

func (o *RecordString) New() string {
	return uuid.New().String()
}
func (o *RecordString) Get() string {
	return o.id
}
func (o *RecordString) IsSet() bool {
	return o.id != ""
}
func (o *RecordString) Set(id string) {
	if o.id != "" {
		return
	}
	o.id = id
}

var _ es.AggregateRoot[any] = (*Record[any])(nil)
var _ es.SnapshotAggregate[any] = (*Record[any])(nil)

var _ es.AggregateID[uuid.UUID] = (*RecordUUID)(nil)
var _ es.AggregateID[int] = (*RecordInt)(nil)
var _ es.AggregateID[string] = (*RecordString)(nil)

func NewRecord[K comparable](id es.AggregateID[K]) *Record[K] {
	return &Record[K]{
		Aggregate: es.NewAggregate(id),
	}
}

func CreateRecord[K comparable](id es.AggregateID[K], text string, number int, timestamp time.Time) (*Record[K], error) {
	record := NewRecord[K](id)

	return record, record.TrackChange(record, &RecordCreated{
		Text:      text,
		Number:    number,
		Timestamp: timestamp,
	})
}

func (r *Record[K]) UpdateText(text string) error {
	return r.TrackChange(r, &RecordTextUpdated{
		Text: text,
	})
}

func (r *Record[K]) UpdateNumber(number int) error {
	return r.TrackChange(r, &RecordNumberUpdated{
		Number: number,
	})
}

func (r *Record[K]) UpdateTimestamp(timestamp time.Time) error {
	return r.TrackChange(r, &RecordTimestampUpdated{
		Timestamp: timestamp,
	})
}

func (r *Record[K]) AggregateType() string {
	return "Record"
}

func (r *Record[K]) ApplyChange(event any) error {
	switch e := event.(type) {
	case *RecordCreated:
		r.Text = e.Text
		r.Number = e.Number
		r.Timestamp = e.Timestamp
	case *RecordTextUpdated:
		r.Text = e.Text
	case *RecordNumberUpdated:
		r.Number = e.Number
	case *RecordTimestampUpdated:
		r.Timestamp = e.Timestamp
	}
	return nil
}

func (r *Record[K]) CreateSnapshot() any {
	return &RecordSnapshot{
		Text:      r.Text,
		Number:    r.Number,
		Timestamp: r.Timestamp,
	}
}

func (r *Record[K]) ApplySnapshot(snapshot any) error {
	switch s := snapshot.(type) {
	case *RecordSnapshot:
		r.Text = s.Text
		r.Number = s.Number
		r.Timestamp = s.Timestamp
	}
	return nil
}

func registerTypes(reg envelope.Registry) error {
	return reg.Register(
		RecordCreated{},
		RecordTextUpdated{},
		RecordNumberUpdated{},
		RecordTimestampUpdated{},
		RecordSnapshot{},
	)
}
