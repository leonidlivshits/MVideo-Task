package service

import (
	"context"
	"time"

	"mvideo-task/internal/domain"
)

type HistoryFilter struct {
	GoodIDs []domain.GoodID
	From    time.Time
	To      time.Time
}

type PriceRepository interface {
	InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error)
	GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error)
	GetHistory(ctx context.Context, filter HistoryFilter) ([]domain.PricePoint, error)
}

type PriceService struct {
	repository PriceRepository
}

func NewPriceService(repository PriceRepository) *PriceService {
	return &PriceService{repository: repository}
}

func (s *PriceService) SetPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	return s.repository.InsertPrices(ctx, prices)
}

func (s *PriceService) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	return s.repository.GetPricesAt(ctx, goodIDs, at)
}

func (s *PriceService) GetHistory(ctx context.Context, filter HistoryFilter) ([]domain.PricePoint, error) {
	return s.repository.GetHistory(ctx, filter)
}
