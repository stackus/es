package es

import (
	"fmt"
)

type ErrUnregisteredEvent string

type ErrUnknownAggregateChangeType struct {
	AggregateType string
	Change        EventPayload
}

type ErrUnregisteredSnapshot string

type ErrUnknownAggregateSnapshotType struct {
	AggregateType string
	Snapshot      SnapshotPayload
}

type ErrAggregateVersionConflict struct {
	AggregateID      any
	AggregateType    string
	AggregateVersion int
}

func (e ErrUnregisteredEvent) Error() string {
	return fmt.Sprintf("unregistered event: %s", string(e))
}

func (e ErrAggregateVersionConflict) Error() string {
	return fmt.Sprintf("aggregate version conflict: %s:%s version:%d", e.AggregateType, e.AggregateID, e.AggregateVersion)
}

func (e ErrUnknownAggregateChangeType) Error() string {
	return fmt.Sprintf("unknown aggregate change type: %s:%s", e.AggregateType, e.Change.Kind())
}

func (e ErrUnregisteredSnapshot) Error() string {
	return fmt.Sprintf("unregistered snapshot: %s", string(e))
}

func (e ErrUnknownAggregateSnapshotType) Error() string {
	return fmt.Sprintf("unknown aggregate snapshot type: %s:%s", e.AggregateType, e.Snapshot.Kind())
}
