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

func NewPgxPool(connection ConnectionString) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), string(connection))
	if err != nil {
		return nil, err
	}
	err = conn.Ping(context.Background())
	if err != nil {
		return nil, err
	}
	// Extract the domain/host from the connection string.
	u, parseErr := url.Parse(string(connection))
	if parseErr != nil {
		glog.Warningf("Failed to parse connection string: %v", parseErr)
	} else {
		glog.Infof("Connected to %s", u.Host)
	}
	return conn, nil
}

func NewDB(conn *pgxpool.Pool) (DB, error) {
	return DB{
		conn:    conn,
		queries: sqlc.New(conn),
	}, nil
}
