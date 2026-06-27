package service

import (
	"context"
	"fmt"
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

type PricesAtResult struct {
	At    time.Time
	Items []domain.PriceAt
}

type PriceService struct {
	repository PriceRepository
}

func NewPriceService(repository PriceRepository) *PriceService {
	return &PriceService{repository: repository}
}

func (s *PriceService) SetPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	if len(prices) == 0 {
		return nil, ErrEmptyPrices
	}

	for _, price := range prices {
		if err := validateGoodID(price.GoodID); err != nil {
			return nil, err
		}
		if price.Price <= 0 {
			return nil, fmt.Errorf("%w: %d", ErrInvalidPrice, price.Price)
		}
	}

	return s.repository.InsertPrices(ctx, prices)
}

func (s *PriceService) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) (PricesAtResult, error) {
	if len(goodIDs) == 0 {
		return PricesAtResult{}, ErrEmptyGoodIDs
	}

	if err := validateGoodIDs(goodIDs); err != nil {
		return PricesAtResult{}, err
	}

	if at.IsZero() {
		at = time.Now()
	}

	prices, err := s.repository.GetPricesAt(ctx, goodIDs, at)
	if err != nil {
		return PricesAtResult{}, err
	}

	return PricesAtResult{
		At:    at,
		Items: prices,
	}, nil
}

func (s *PriceService) GetHistory(ctx context.Context, filter HistoryFilter) ([]domain.PricePoint, error) {
	if filter.From.IsZero() || filter.To.IsZero() || !filter.From.Before(filter.To) {
		return nil, ErrInvalidPeriod
	}

	if err := validateGoodIDs(filter.GoodIDs); err != nil {
		return nil, err
	}

	return s.repository.GetHistory(ctx, filter)
}

func validateGoodIDs(goodIDs []domain.GoodID) error {
	for _, goodID := range goodIDs {
		if err := validateGoodID(goodID); err != nil {
			return err
		}
	}

	return nil
}

func validateGoodID(goodID domain.GoodID) error {
	if goodID <= 0 {
		return fmt.Errorf("%w: %d", ErrInvalidGoodID, goodID)
	}

	return nil
}
