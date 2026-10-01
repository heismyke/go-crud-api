package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	// Registers the "pgx" driver with database/sql. The underscore import
	// means "run this package's init(), I won't call it directly".
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Open connects to Postgres and verifies the connection with a ping.
//
// *sql.DB is NOT a single connection — it's a pool. It's safe for concurrent
// use and meant to be created once and shared for the life of the program.
func Open(dsn string) (*sql.DB, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}

	// A small pool makes it easy to watch connections in pg_stat_activity.
	db.SetMaxOpenConns(3)
	db.SetMaxIdleConns(3)
	db.SetConnMaxLifetime(30 * time.Minute)

	// sql.Open doesn't actually connect. Ping forces a real round trip.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return db, nil
}
