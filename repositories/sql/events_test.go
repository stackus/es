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

type eventSuite[K comparable] struct {
	suite.Suite
	ctx       context.Context
	setupFn   func(ctx context.Context, files []string) (*sql.DB, func() error, error)
	cleanupFn func() error
	db        *sql.DB
	repo      es.EventRepository[K]
	store     es.AggregateStore[K]
	idFactory func() es.AggregateID[K]
	files     []string
}

func (s *eventSuite[K]) SetupSuite() {
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
	s.repo = essql.NewEventRepository[K](s.db)
	s.store = es.NewEventStore(reg, s.repo)
}

func (s *eventSuite[K]) TearDownSuite() {
	if s.cleanupFn != nil {
		if err := s.cleanupFn(); err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *eventSuite[K]) TestEventRepository_PreLoadError() {
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
	preLoadErr := errors.New("pre-load error")
	m.On("EventsPreLoad", mock.Anything, record).Return(preLoadErr)
	events, err := s.repo.Load(s.ctx, record, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), preLoadErr, err)
	assert.Nil(s.T(), events)

	m.AssertExpectations(s.T())
}

func (s *eventSuite[K]) TestEventRepository_NoEvents() {
	record := NewRecord(s.idFactory())
	record.SetID(s.idFactory().New())
	m := &hookMock[K]{}
	m.On("EventsPreLoad", mock.Anything, record).Return(nil)
	m.On("EventsPostLoad", mock.Anything, record, mock.Anything).Return(nil)
	events, err := s.repo.Load(s.ctx, record, m)
	assert.NoError(s.T(), err)
	assert.Len(s.T(), events, 0)

	m.AssertExpectations(s.T())
}

func (s *eventSuite[K]) TestEventRepository_PostLoadError() {
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
	m.On("EventsPreLoad", mock.Anything, record).Return(nil)
	m.On("EventsPostLoad", mock.Anything, record, mock.Anything).Return(postLoadErr)
	events, err := s.repo.Load(s.ctx, record, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), postLoadErr, err)
	assert.Nil(s.T(), events)

	m.AssertExpectations(s.T())
}

func (s *eventSuite[K]) TestEventStore_Load() {
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

func (s *eventSuite[K]) TestEventRepository_PreSaveError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	preSaveErr := errors.New("pre-save error")
	m.On("EventsPreSave", mock.Anything, record, mock.Anything).Return(preSaveErr)
	err := s.repo.Save(s.ctx, record, []es.Event[K]{}, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), preSaveErr, err)

	m.AssertExpectations(s.T())
}

func (s *eventSuite[K]) TestEventRepository_PostSaveError() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	m := &hookMock[K]{}
	postSaveErr := errors.New("post-save error")
	m.On("EventsPreSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("EventsPostSave", mock.Anything, record, mock.Anything).Return(postSaveErr)
	err := s.repo.Save(s.ctx, record, []es.Event[K]{}, m)
	assert.Error(s.T(), err)
	assert.Equal(s.T(), postSaveErr, err)

	m.AssertExpectations(s.T())
}

func (s *eventSuite[K]) TestEventStore_Save() {
	record, _ := CreateRecord(s.idFactory(), "test", 1, time.Time{})
	_ = record.UpdateText("updated")
	_ = record.UpdateNumber(2)
	_ = record.UpdateTimestamp(time.Now())

	var savedEvents []es.Event[K]
	eventsSaveHook := es.EventsPostSaveHook(func(ctx context.Context, aggregate es.Aggregate[K], events []es.Event[K]) error {
		assert.Len(s.T(), events, 4)
		assert.Equal(s.T(), record.AggregateID(), events[0].AggregateID)
		assert.Equal(s.T(), record.AggregateType(), events[0].AggregateType)
		savedEvents = make([]es.Event[K], len(events))
		copy(savedEvents, events)
		return nil
	})
	m := &hookMock[K]{}
	m.On("EventsPreSave", mock.Anything, record, mock.Anything).Return(nil)
	m.On("EventsPostSave", mock.Anything, record, mock.Anything).Return(nil)
	if err := s.store.Save(s.ctx, record, eventsSaveHook, m); err != nil {
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

	m.AssertExpectations(s.T())
}

func TestEventSuite(t *testing.T) {
	t.Run("postgres:string", func(t *testing.T) {
		suite.Run(t, &eventSuite[string]{
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
		suite.Run(t, &eventSuite[int]{
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
		suite.Run(t, &eventSuite[uuid.UUID]{
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
		suite.Run(t, &eventSuite[string]{
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
		suite.Run(t, &eventSuite[int]{
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
		suite.Run(t, &eventSuite[uuid.UUID]{
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
