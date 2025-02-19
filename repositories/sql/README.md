# SQL Repositories for Event Sourcing

A set of SQL, `database/sql`, repositories for event sourcing. They're designed to be database-agnostic, and are tested with the following databases:

- PostgreSQL
- SQLite

## Schema

The schema for the two tables can be found in the `schema` directory. You should copy, modify as necessary, and then add them to whatever migration tool you're using.

These schemas should be considered working examples and may not be suitable for production use. 

### Generic Aggregate ID

For each database, the aggregate ID is stored as a blob, or byte array. This is to allow for any type of ID to be used. Using integers, strings, and UUIDs are all possible, but performance may suffer.

For performance reasons, it is recommended to use a column type aligned with the type of ID you're using. For example, if you're using UUIDs, use the `uuid` type in PostgreSQL.
