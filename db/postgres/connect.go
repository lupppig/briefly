package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
)

type PostgresDB struct {
	Conn *pgx.Conn
}

func ConnectPostgres(ctx context.Context, dbUrl string) (*PostgresDB, error) {
	conn, err := pgx.Connect(ctx, dbUrl)

	if err != nil {
		return nil, fmt.Errorf("unable to connect to database %w", err)
	}

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database :%w", err)
	}

	log.Println("connected to database succesful")
	pg := &PostgresDB{Conn: conn}
	return pg, nil
}
