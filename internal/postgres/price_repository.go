package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"mvideo-task/internal/domain"
	"mvideo-task/internal/service"
)

type PriceRepository struct {
	pool *pgxpool.Pool
}

func NewPriceRepository(pool *pgxpool.Pool) *PriceRepository {
	return &PriceRepository{pool: pool}
}

func (r *PriceRepository) InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	return nil, nil
}

func (r *PriceRepository) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	return nil, nil
}

func (r *PriceRepository) GetHistory(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error) {
	return nil, nil
}
