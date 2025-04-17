package memory

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stackus/es"
)

func TestEventRepository_Load(t *testing.T) {
	type fields struct {
		streams map[string]map[string][]es.Event[string]
	}
	type args struct {
		ctx       context.Context
		aggregate es.AggregateRoot[string]
	}

	// to simply testing the time values
	now := time.Now()

	// simplify the error testing
	errTest := fmt.Errorf("test error")

	tests := map[string]struct {
		fields     fields
		args       args
		setupHooks func(hooks *hookMock[string])
		want       []es.Event[string]
		wantErr    bool
	}{
		"should load events": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{
					"Record": {
						"1": {
							{
								AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
								EventData:  []byte("test"),
								OccurredAt: now,
							},
						},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreLoad", context.Background(), mock.Anything).Return(nil)
				hooks.On("EventsPostLoad", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			want: []es.Event[string]{
				{
					AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test", EventData: []byte("test"),
					OccurredAt: now,
				},
			},
			wantErr: false,
		},
		"should return nil if no events found": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{
					"Record": {},
				},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreLoad", context.Background(), mock.Anything).Return(nil)
				hooks.On("EventsPostLoad", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		"should return nil if aggregate version is greater than the number of events": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{
					"Record": {
						"1": {
							{
								AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
								EventData:  []byte("test"),
								OccurredAt: now,
							},
						},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					_ = a.UpdateText("updated")
					a.SetID("1")

					return a
				}(),
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreLoad", context.Background(), mock.Anything).Return(nil)
				hooks.On("EventsPostLoad", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			want:    nil,
			wantErr: false,
		},
		"should return an error if pre-load hook fails": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{
					"Record": {
						"1": {
							{
								AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
								EventData:  []byte("test"),
								OccurredAt: now,
							},
						},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreLoad", context.Background(), mock.Anything).Return(errTest)
			},
			want:    nil,
			wantErr: true,
		},
		"should return an error if post-load hook fails": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{
					"Record": {
						"1": {
							{
								AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
								EventData:  []byte("test"),
								OccurredAt: now,
							},
						},
					},
				},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreLoad", context.Background(), mock.Anything).Return(nil)
				hooks.On("EventsPostLoad", context.Background(), mock.Anything, mock.Anything).Return(errTest)
			},
			want:    nil,
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &eventRepository[string]{
				streams: tt.fields.streams,
			}
			hooks := &hookMock[string]{}
			if tt.setupHooks != nil {
				tt.setupHooks(hooks)
			}
			got, err := r.Load(tt.args.ctx, tt.args.aggregate, hooks)
			if (err != nil) != tt.wantErr {
				t.Errorf("eventRepository.Load() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("eventRepository.Load() = %v, want %v", got, tt.want)
			}

			hooks.AssertExpectations(t)
		})
	}
}

func TestEventRepository_Save(t *testing.T) {
	type fields struct {
		streams map[string]map[string][]es.Event[string]
	}
	type args struct {
		ctx       context.Context
		aggregate es.AggregateRoot[string]
		events    []es.Event[string]
	}

	// to simply testing the time values
	now := time.Now()

	// simplify the error testing
	errTest := fmt.Errorf("test error")

	tests := map[string]struct {
		fields     fields
		args       args
		setupHooks func(hooks *hookMock[string])
		wantErr    bool
	}{
		"should save events": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				events: []es.Event[string]{
					{
						AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
						EventData:  []byte("test"),
						OccurredAt: now,
					},
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreSave", context.Background(), mock.Anything, mock.Anything).Return(nil)
				hooks.On("EventsPostSave", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: false,
		},
		"should return an error if pre-save hook fails": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				events: []es.Event[string]{
					{
						AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
						EventData:  []byte("test"),
						OccurredAt: now,
					},
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreSave", context.Background(), mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
		"should return an error if post-save hook fails": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				events: []es.Event[string]{
					{
						AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, EventType: "test",
						EventData:  []byte("test"),
						OccurredAt: now,
					},
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreSave", context.Background(), mock.Anything, mock.Anything).Return(nil)
				hooks.On("EventsPostSave", context.Background(), mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
		"should return an error if the event version does not match the aggregate version": {
			fields: fields{
				streams: map[string]map[string][]es.Event[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				events: []es.Event[string]{
					{
						AggregateID: "1", AggregateType: "Record", AggregateVersion: 2, EventType: "test",
						EventData:  []byte("test"),
						OccurredAt: now,
					},
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("EventsPreSave", context.Background(), mock.Anything, mock.Anything).Return(nil)
			},
			wantErr: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			r := &eventRepository[string]{
				streams: tt.fields.streams,
			}
			hooks := &hookMock[string]{}
			if tt.setupHooks != nil {
				tt.setupHooks(hooks)
			}
			if err := r.Save(tt.args.ctx, tt.args.aggregate, tt.args.events, hooks); (err != nil) != tt.wantErr {
				t.Errorf("eventRepository.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			hooks.AssertExpectations(t)
		})
	}
}
