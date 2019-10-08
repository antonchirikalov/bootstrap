package postgres

import (
	"context"

	"github.com/jackc/pgx"
)

// DB is a configured instance of a postgres database
type DB struct {
	Pool *pgx.ConnPool
}

// Connect creates the initial context for operating with the PostgresDB
//
// Return Values:
//     1st: An error representing failure to connect
func Connect(config *pgx.ConnPoolConfig) (*DB, error) {
	pool, err := pgx.NewConnPool(*config)
	if err != nil {
		return nil, err
	}

	return &DB{pool}, nil
}

// Ping verifies a connection to the database is still alive, establishing a connection if necessary
//
// Return Values:
//     1st: An error representing failure to connect
func (db *DB) Ping() error {
	conn, err := db.Pool.Acquire()
	if err != nil {
		return err
	}
	defer db.Pool.Release(conn)

	return conn.Ping(context.Background())
}

// Close closes the database, releasing any open resources
// It is rare to Close a DB, as the DB handle is meant to be long-lived and shared between many goroutines
//
// Return Values:
//     1st: An error representing failure to close the connection
func (db *DB) Close() {
	db.Pool.Close()
}
