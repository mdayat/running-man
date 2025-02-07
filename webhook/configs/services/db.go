package services

import (
	"context"
	"webhook/repository"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	DB      *pgxpool.Pool
	Queries *repository.Queries
)

func NewDB(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	var err error
	DB, err = pgxpool.New(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	Queries = repository.New(DB)

	return DB, err
}
