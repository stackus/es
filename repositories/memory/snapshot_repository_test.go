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

func TestSnapshotRepository_Load(t *testing.T) {
	type fields struct {
		snapshots map[string]map[string]es.Snapshot[string]
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
		want       *es.Snapshot[string]
		wantErr    bool
	}{
		"should load snapshot": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{
					"Record": {
						"1": {
							AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, SnapshotData: []byte("test"), CreatedAt: now,
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
				hooks.On("SnapshotPreLoad", mock.Anything, mock.Anything).Return(nil)
				hooks.On("SnapshotPostLoad", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
			want: &es.Snapshot[string]{
				AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, SnapshotData: []byte("test"), CreatedAt: now,
			},
		},
		"should return nil if snapshot not found": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{},
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
				hooks.On("SnapshotPreLoad", mock.Anything, mock.Anything).Return(nil)
			},
			want: nil,
		},
		"should return an error if pre-load hook fails": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{},
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
				hooks.On("SnapshotPreLoad", mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
		"should return an error if post-load hook fails": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{
					"Record": {
						"1": {
							AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, SnapshotData: []byte("test"), CreatedAt: now,
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
				hooks.On("SnapshotPreLoad", mock.Anything, mock.Anything).Return(nil)
				hooks.On("SnapshotPostLoad", mock.Anything, mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &snapshotRepository[string]{
				snapshots: tt.fields.snapshots,
			}
			hooks := &hookMock[string]{}
			if tt.setupHooks != nil {
				tt.setupHooks(hooks)
			}

			snapshot, err := repo.Load(tt.args.ctx, tt.args.aggregate, hooks)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotRepository.Load() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if !reflect.DeepEqual(snapshot, tt.want) {
					t.Errorf("SnapshotRepository.Load() = %v, want %v", snapshot, tt.want)
				}
			}
		})
	}
}

func TestSnapshotRepository_Save(t *testing.T) {
	type fields struct {
		snapshots map[string]map[string]es.Snapshot[string]
	}
	type args struct {
		ctx       context.Context
		aggregate es.AggregateRoot[string]
		snapshot  es.Snapshot[string]
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
		"should save snapshot": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{
					"Record": {
						"1": {
							AggregateID: "1", AggregateType: "Record", AggregateVersion: 1, SnapshotData: []byte("test"), CreatedAt: now,
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
				snapshot: es.Snapshot[string]{
					AggregateID: "1", AggregateType: "Record", AggregateVersion: 2, SnapshotData: []byte("test2"), CreatedAt: now,
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("SnapshotPreSave", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				hooks.On("SnapshotPostSave", mock.Anything, mock.Anything, mock.Anything).Return(nil)
			},
		},
		"should return an error if pre-save hook fails": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				snapshot: es.Snapshot[string]{
					AggregateID: "1", AggregateType: "Record", AggregateVersion: 2, SnapshotData: []byte("test2"), CreatedAt: now,
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("SnapshotPreSave", mock.Anything, mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
		"should return an error if post-save hook fails": {
			fields: fields{
				snapshots: map[string]map[string]es.Snapshot[string]{},
			},
			args: args{
				ctx: context.Background(),
				aggregate: func() es.AggregateRoot[string] {
					a := NewRecord[string](&RecordString{})
					a.SetID("1")

					return a
				}(),
				snapshot: es.Snapshot[string]{
					AggregateID: "1", AggregateType: "Record", AggregateVersion: 2, SnapshotData: []byte("test2"), CreatedAt: now,
				},
			},
			setupHooks: func(hooks *hookMock[string]) {
				hooks.On("SnapshotPreSave", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				hooks.On("SnapshotPostSave", mock.Anything, mock.Anything, mock.Anything).Return(errTest)
			},
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			repo := &snapshotRepository[string]{
				snapshots: tt.fields.snapshots,
			}
			hooks := &hookMock[string]{}
			if tt.setupHooks != nil {
				tt.setupHooks(hooks)
			}

			err := repo.Save(tt.args.ctx, tt.args.aggregate, tt.args.snapshot, hooks)
			if (err != nil) != tt.wantErr {
				t.Errorf("SnapshotRepository.Save() error = %v, wantErr %v", err, tt.wantErr)
			}

			hooks.AssertExpectations(t)
		})
	}
}
