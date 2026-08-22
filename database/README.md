# Database

The data layer, in three passes. [`models/`](models/) holds the domain types — plain
Go structs, one per table. [`stores/`](stores/) is the Postgres access layer over those
types, built on [astql](https://github.com/zoobz-io/astql) + sqlx. [`migrations/`](migrations/)
is the schema itself: goose SQL files that create the tables the models map to and the
indexes the stores rely on.

A store reads and writes a model; a model maps 1:1 to a table a migration created. For
how these types relate as a domain — entitlements, the one-round-trip resolver — see the
[root README](../README.md).
