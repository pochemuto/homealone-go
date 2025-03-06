package db

import (
	"context"
	"net/url"

	"github.com/golang/glog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pochemuto/homealone-go/internal/sqlc"
)

type DB struct {
	conn    *pgxpool.Pool
	queries *sqlc.Queries
}

type ConnectionString string

func NewPgxPool(connectionString ConnectionString) (*pgxpool.Pool, error) {
	// Creates a new pgx pool.
	conn, err := pgxpool.New(context.Background(), string(connectionString))
	if err != nil {
		return nil, err
	}
	// Check connection.
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	// Extract the domain/host from the connection string.
	u, parseErr := url.Parse(string(connectionString))
	if parseErr != nil {
		glog.Warningf("Failed to parse connection string: %v", parseErr)
	} else {
		glog.Infof("Connected to %s, db %s", u.Host, u.Path[1:])
	}
	return conn, nil
}

func NewDB(conn *pgxpool.Pool) (DB, error) {
	// Creates a new DB instance.
	return DB{
		conn:    conn,
		queries: sqlc.New(conn),
	}, nil
}
