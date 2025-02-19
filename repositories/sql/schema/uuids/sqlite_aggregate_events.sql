-- SQLite events schema

-- Create the events table with UUID aggregate IDs
CREATE TABLE aggregate_events (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	aggregate_id      BLOB    NOT NULL,
	aggregate_type    TEXT    NOT NULL,
	aggregate_version INTEGER NOT NULL,
	event_type        TEXT    NOT NULL,
	event_data        BLOB    NOT NULL,
	occurred_at       DATE    NOT NULL,
	created_at        DATE    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create an index on the aggregate_it/type column ordered aggregate_version
CREATE UNIQUE INDEX idx_aggregate_events_type_version ON aggregate_events (aggregate_id, aggregate_type, aggregate_version);
