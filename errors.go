package es

import (
	"fmt"
)

type ErrUnknownAggregateChangeType struct {
	AggregateType string
	Change        any
}

type ErrUnknownAggregateSnapshotType struct {
	AggregateType string
	Snapshot      any
}

type ErrAggregateVersionConflict struct {
	AggregateID      any
	AggregateType    string
	AggregateVersion int
}

func (e ErrAggregateVersionConflict) Error() string {
	return fmt.Sprintf("aggregate version conflict: %s:%s version:%d", e.AggregateType, e.AggregateID, e.AggregateVersion)
}

func (e ErrUnknownAggregateChangeType) Error() string {
	return fmt.Sprintf("unknown aggregate change type: %s:%T", e.AggregateType, e.Change)
}

func (e ErrUnknownAggregateSnapshotType) Error() string {
	return fmt.Sprintf("unknown aggregate snapshot type: %s:%T", e.AggregateType, e.Snapshot)
}
