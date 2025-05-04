package memory

import (
	"time"

	"github.com/google/uuid"

	"github.com/stackus/es"
)

type (
	Record[K comparable] struct {
		es.Aggregate[K]
		Text      string
		Number    int
		Timestamp time.Time
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

func (RecordCreated) Kind() string {
	return "RecordCreated"
}
func (RecordTextUpdated) Kind() string {
	return "RecordTextUpdated"
}
func (RecordNumberUpdated) Kind() string {
	return "RecordNumberUpdated"
}
func (RecordTimestampUpdated) Kind() string {
	return "RecordTimestampUpdated"
}
func (RecordSnapshot) Kind() string {
	return "RecordSnapshot"
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

func (r *Record[K]) ApplyChange(event es.EventPayload) error {
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

func (r *Record[K]) CreateSnapshot() es.SnapshotPayload {
	return &RecordSnapshot{
		Text:      r.Text,
		Number:    r.Number,
		Timestamp: r.Timestamp,
	}
}

func (r *Record[K]) ApplySnapshot(snapshot es.SnapshotPayload) error {
	switch s := snapshot.(type) {
	case *RecordSnapshot:
		r.Text = s.Text
		r.Number = s.Number
		r.Timestamp = s.Timestamp
	}
	return nil
}

// func registerTypes(reg envelope.Registry) error {
// 	return reg.Register(
// 		RecordCreated{},
// 		RecordTextUpdated{},
// 		RecordNumberUpdated{},
// 		RecordTimestampUpdated{},
// 		RecordSnapshot{},
// 	)
// }
