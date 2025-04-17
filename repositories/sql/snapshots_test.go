package sql_test

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stackus/envelope"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/stackus/es"
	essql "github.com/stackus/es/repositories/sql"
)

type snapshotSuite[K comparable] struct {
	suite.Suite
	ctx       context.Context
	setupFn   func(ctx context.Context, files []string) (*sql.DB, func() error, error)
	cleanupFn func() error
	db        *sql.DB
	repo      es.SnapshotRepository[K]
	store     es.AggregateStore[K]
	idFactory func() es.AggregateID[K]
	files     []string
}

func (s *snapshotSuite[K]) SetupSuite() {
	s.ctx = context.Background()
	db, cleanup, err := s.setupFn(s.ctx, s.files)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	s.db = db
	s.cleanupFn = cleanup
	reg := envelope.NewRegistry()
	err = registerTypes(reg)
	s.repo = essql.NewSnapshotRepository[K](s.db)
	s.store = es.NewSnapshotStore(
		reg,
		es.NewEventStore(reg, essql.NewEventRepository[K](s.db)),
		s.repo,
		es.NewFrequencySnapshotStrategy[K](2),
	)
}

func (s *snapshotSuite[K]) TearDownSuite() {
	if s.cleanupFn != nil {
		if err := s.cleanupFn(); err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *snapshotSuite[K]) TestSnapshotRepository_PreLoadError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	preLoadErr := errors.New("pre-load error")
	m.On("SnapshotPreLoad", mock.Anything, record).Return(preLoadErr)
	_, err := s.repo.Load(s.ctx, record, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), preLoadErr, err)
	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotRepository_NoSnapshot() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	m.On("SnapshotPreLoad", mock.Anything, record).Return(nil)
	m.On("SnapshotPostLoad", mock.Anything, record, mock.Anything).Return(nil)
	_, err := s.repo.Load(s.ctx, record, m)
	assert.NoError(s.T(), err)
	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotRepository_PostLoadError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	if err := s.store.Save(s.ctx, record); err != nil {
		s.T().Fatal(err)
		return
	}
	assert.Equal(s.T(), 4, record.AggregateVersion())

	m := &hookMock[K]{}
	postLoadErr := errors.New("post-load error")
	m.On("SnapshotPreLoad", mock.Anything, record).Return(nil)
	m.On("SnapshotPostLoad", mock.Anything, record, mock.Anything).Return(postLoadErr)
	_, err := s.repo.Load(s.ctx, record, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), postLoadErr, err)
	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotStore_Load() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	if err := s.store.Save(s.ctx, record); err != nil {
		s.T().Fatal(err)
		return
	}
	assert.Equal(s.T(), 4, record.AggregateVersion())

	loadedRecord := NewRecord(s.idFactory())
	loadedRecord.SetID(record.AggregateID())
	m := &hookMock[K]{}
	m.On("EventsPreLoad", mock.Anything, loadedRecord).Return(nil)
	m.On("EventsPostLoad", mock.Anything, loadedRecord, mock.Anything).Return(nil)
	m.On("SnapshotPreLoad", mock.Anything, loadedRecord).Return(nil)
	m.On("SnapshotPostLoad", mock.Anything, loadedRecord, mock.Anything).Return(nil)
	err := s.store.Load(s.ctx, loadedRecord, m)
	if err != nil {
		s.T().Fatal(err)
		return
	}

	assert.Equal(s.T(), record.AggregateVersion(), loadedRecord.AggregateVersion())
	assert.Equal(s.T(), record.Text, loadedRecord.Text)
	assert.Equal(s.T(), record.Number, loadedRecord.Number)
	assert.Equal(s.T(), record.Timestamp.Format(time.RFC3339), loadedRecord.Timestamp.Format(time.RFC3339))

	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotRepository_PreSaveError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	preSaveErr := errors.New("pre-save error")
	m.On("SnapshotPreSave", mock.Anything, record, mock.Anything).Return(preSaveErr)
	err := s.repo.Save(s.ctx, record, es.Snapshot[K]{}, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), preSaveErr, err)
	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotRepository_PostSaveError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	postSaveErr := errors.New("post-save error")
	m.On("SnapshotPreSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("SnapshotPostSave", mock.Anything, record, mock.Anything).Return(postSaveErr)
	snapshot := es.Snapshot[K]{
		AggregateID:      record.AggregateID(),
		AggregateType:    record.AggregateType(),
		AggregateVersion: record.AggregateVersion(),
		SnapshotType:     "test",
		SnapshotData:     []byte("test"),
		CreatedAt:        time.Now(),
	}
	err := s.repo.Save(s.ctx, record, snapshot, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), postSaveErr, err)
	m.AssertExpectations(s.T())
}

func (s *snapshotSuite[K]) TestSnapshotStore_Save() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	var savedEvents []es.Event[K]
	eventsSaveHook := es.EventsPostSaveHook(func(ctx context.Context, aggregate es.AggregateRoot[K], events []es.Event[K]) error {
		assert.Len(s.T(), events, 4)
		assert.Equal(s.T(), record.AggregateID(), events[0].AggregateID)
		assert.Equal(s.T(), record.AggregateType(), events[0].AggregateType)
		savedEvents = make([]es.Event[K], len(events))
		copy(savedEvents, events)
		return nil
	})
	var savedSnapshot es.Snapshot[K]
	snapshotSaveHook := es.SnapshotPostSaveHook(func(ctx context.Context, aggregate es.AggregateRoot[K], snapshot es.Snapshot[K]) error {
		assert.Equal(s.T(), record.AggregateID(), snapshot.AggregateID)
		assert.Equal(s.T(), record.AggregateType(), snapshot.AggregateType)
		savedSnapshot = snapshot
		return nil
	})
	m := &hookMock[K]{}
	m.On("EventsPreSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("EventsPostSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("SnapshotPreSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("SnapshotPostSave", mock.Anything, record, mock.Anything).Return(nil)
	if err := s.store.Save(s.ctx, record, eventsSaveHook, snapshotSaveHook, m); err != nil {
		s.T().Fatal(err)
		return
	}
	assert.Equal(s.T(), 4, record.AggregateVersion())

	// select directly from the database to verify the data was stored correctly
	rows, err := s.db.QueryContext(s.ctx, "SELECT aggregate_type, aggregate_version, event_type, event_data FROM aggregate_events WHERE aggregate_id = $1 ORDER BY aggregate_version", record.AggregateID())
	if err != nil {
		s.T().Fatal(err)
		return
	}
	defer rows.Close()
	var count int

	for rows.Next() {
		var aggregateType string
		var aggregateVersion int
		var eventType string
		var eventData []byte
		count++
		if err := rows.Scan(&aggregateType, &aggregateVersion, &eventType, &eventData); err != nil {
			s.T().Fatal(err)
			return
		}
		assert.Equal(s.T(), record.AggregateType(), aggregateType)
		assert.Equal(s.T(), savedEvents[0].AggregateVersion, aggregateVersion)
		assert.Equal(s.T(), savedEvents[0].EventType, eventType)
		assert.Equal(s.T(), savedEvents[0].EventData, eventData)
		savedEvents = savedEvents[1:]
	}
	assert.Equal(s.T(), 4, count)

	// select directly from the database to verify the snapshot was stored correctly
	row := s.db.QueryRowContext(s.ctx, "SELECT aggregate_type, aggregate_version, snapshot_data FROM aggregate_snapshots WHERE aggregate_id = $1", record.AggregateID())
	var aggregateType string
	var aggregateVersion int
	var snapshotData []byte
	if err := row.Scan(&aggregateType, &aggregateVersion, &snapshotData); err != nil {
		s.T().Fatal(err)
		return
	}
	assert.Equal(s.T(), savedSnapshot.AggregateType, aggregateType)
	assert.Equal(s.T(), savedSnapshot.AggregateVersion, aggregateVersion)
	assert.Equal(s.T(), savedSnapshot.SnapshotData, snapshotData)

	m.AssertExpectations(s.T())
}

func TestSnapshotSuite(t *testing.T) {
	t.Run("postgres:string", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[string]{
			setupFn: usePostgres,
			idFactory: func() es.AggregateID[string] {
				return &RecordString{}
			},
			files: []string{
				filepath.Join("schema", "strings", "postgres_aggregate_events.sql"),
				filepath.Join("schema", "strings", "postgres_aggregate_snapshots.sql"),
			},
		})
	})
	t.Run("postgres:int", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[int]{
			setupFn: usePostgres,
			idFactory: func() es.AggregateID[int] {
				return &RecordInt{}
			},
			files: []string{
				filepath.Join("schema", "ints", "postgres_aggregate_events.sql"),
				filepath.Join("schema", "ints", "postgres_aggregate_snapshots.sql"),
			},
		})
	})
	t.Run("postgres:uuid", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[uuid.UUID]{
			setupFn: usePostgres,
			idFactory: func() es.AggregateID[uuid.UUID] {
				return &RecordUUID{}
			},
			files: []string{
				filepath.Join("schema", "uuids", "postgres_aggregate_events.sql"),
				filepath.Join("schema", "uuids", "postgres_aggregate_snapshots.sql"),
			},
		})
	})
	t.Run("sqlite:string", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[string]{
			setupFn: useSqlite,
			idFactory: func() es.AggregateID[string] {
				return &RecordString{}
			},
			files: []string{
				filepath.Join("schema", "strings", "sqlite_aggregate_events.sql"),
				filepath.Join("schema", "strings", "sqlite_aggregate_snapshots.sql"),
			},
		})
	})
	t.Run("sqlite:int", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[int]{
			setupFn: useSqlite,
			idFactory: func() es.AggregateID[int] {
				return &RecordInt{}
			},
			files: []string{
				filepath.Join("schema", "ints", "sqlite_aggregate_events.sql"),
				filepath.Join("schema", "ints", "sqlite_aggregate_snapshots.sql"),
			},
		})
	})
	t.Run("sqlite:uuid", func(t *testing.T) {
		suite.Run(t, &snapshotSuite[uuid.UUID]{
			setupFn: useSqlite,
			idFactory: func() es.AggregateID[uuid.UUID] {
				return &RecordUUID{}
			},
			files: []string{
				filepath.Join("schema", "uuids", "sqlite_aggregate_events.sql"),
				filepath.Join("schema", "uuids", "sqlite_aggregate_snapshots.sql"),
			},
		})
	})
}
