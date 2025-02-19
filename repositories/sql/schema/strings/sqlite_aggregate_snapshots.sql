-- SQLite snapshots schema

-- Create the snapshots table with string aggregate IDs
CREATE TABLE aggregate_snapshots (
	id                INTEGER PRIMARY KEY AUTOINCREMENT,
	aggregate_id      TEXT    NOT NULL,
	aggregate_type    TEXT    NOT NULL,
	aggregate_version INTEGER NOT NULL,
	snapshot_type     TEXT    NOT NULL,
	snapshot_data     BLOB    NOT NULL,
	created_at        DATE    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create an index on the aggregate_id/type column ordered aggregate_version
CREATE UNIQUE INDEX idx_aggregate_snapshots_type_version ON aggregate_snapshots (aggregate_id, aggregate_type, aggregate_version);
