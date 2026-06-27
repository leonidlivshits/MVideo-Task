package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"mvideo-task/internal/domain"
)

type fakePriceRepository struct {
	insertedPrices []domain.PricePoint
	pricesAtIDs    []domain.GoodID
	pricesAtTime   time.Time
	historyFilter  HistoryFilter
	err            error
}

func (r *fakePriceRepository) InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	if r.err != nil {
		return nil, r.err
	}

	r.insertedPrices = prices
	return prices, nil
}

func (r *fakePriceRepository) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	if r.err != nil {
		return nil, r.err
	}

	r.pricesAtIDs = goodIDs
	r.pricesAtTime = at
	return []domain.PriceAt{}, nil
}

func (r *fakePriceRepository) GetHistory(ctx context.Context, filter HistoryFilter) ([]domain.PricePoint, error) {
	if r.err != nil {
		return nil, r.err
	}

	r.historyFilter = filter
	return []domain.PricePoint{}, nil
}

// Пустой список цен отклоняется.
func TestSetPricesReturnsErrorForEmptyPrices(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.SetPrices(context.Background(), nil)

	if !errors.Is(err, ErrEmptyPrices) {
		t.Errorf("expected ErrEmptyPrices, got %v", err)
	}
}

// good_id должен быть положительным при установке цен
func TestSetPricesReturnsErrorForInvalidGoodID(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.SetPrices(context.Background(), []domain.PricePoint{
		{GoodID: 0, Price: 100},
	})

	if !errors.Is(err, ErrInvalidGoodID) {
		t.Errorf("expected ErrInvalidGoodID, got %v", err)
	}
}

// Нулевая цена отклоняется
func TestSetPricesReturnsErrorForZeroPrice(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.SetPrices(context.Background(), []domain.PricePoint{
		{GoodID: 1, Price: 0},
	})

	if !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

// Отрицательная цена отклоняется
func TestSetPricesReturnsErrorForNegativePrice(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.SetPrices(context.Background(), []domain.PricePoint{
		{GoodID: 1, Price: -1},
	})

	if !errors.Is(err, ErrInvalidPrice) {
		t.Errorf("expected ErrInvalidPrice, got %v", err)
	}
}

// Валидные цены передаются в репозиторий
func TestSetPricesCallsRepository(t *testing.T) {
	repository := &fakePriceRepository{}
	service := NewPriceService(repository)

	_, err := service.SetPrices(context.Background(), []domain.PricePoint{
		{GoodID: 1, Price: 100},
		{GoodID: 2, Price: 200},
	})

	if err != nil {
		t.Errorf("set prices: %v", err)
		return
	}
	if len(repository.insertedPrices) != 2 {
		t.Errorf("expected 2 prices, got %d", len(repository.insertedPrices))
		return
	}
	if repository.insertedPrices[0].GoodID != 1 || repository.insertedPrices[0].Price != 100 {
		t.Errorf("unexpected first price: %+v", repository.insertedPrices[0])
	}
	if repository.insertedPrices[1].GoodID != 2 || repository.insertedPrices[1].Price != 200 {
		t.Errorf("unexpected second price: %+v", repository.insertedPrices[1])
	}
}

// Проверяет, что сервис возвращает ошибку репозитория.
func TestSetPricesReturnsRepositoryError(t *testing.T) {
	repositoryError := errors.New("repository error")
	service := NewPriceService(&fakePriceRepository{err: repositoryError})

	_, err := service.SetPrices(context.Background(), []domain.PricePoint{
		{GoodID: 1, Price: 100},
	})

	if !errors.Is(err, repositoryError) {
		t.Errorf("expected repository error, got %v", err)
	}
}

// Проверяет, что цены нельзя запросить без товаров
func TestGetPricesAtReturnsErrorForEmptyGoodIDs(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.GetPricesAt(context.Background(), nil, time.Now())

	if !errors.Is(err, ErrEmptyGoodIDs) {
		t.Errorf("expected ErrEmptyGoodIDs, got %v", err)
	}
}

// Проверяет, что good_id должен быть положительным при чтении цен
func TestGetPricesAtReturnsErrorForInvalidGoodID(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.GetPricesAt(context.Background(), []domain.GoodID{1, 0}, time.Now())

	if !errors.Is(err, ErrInvalidGoodID) {
		t.Errorf("expected ErrInvalidGoodID, got %v", err)
	}
}

// Проверяет, что если время не передали, сервис берет текущее время
func TestGetPricesAtUsesCurrentTimeWhenAtIsZero(t *testing.T) {
	repository := &fakePriceRepository{}
	service := NewPriceService(repository)

	result, err := service.GetPricesAt(context.Background(), []domain.GoodID{1}, time.Time{})

	if err != nil {
		t.Errorf("get prices at: %v", err)
		return
	}
	if result.At.IsZero() {
		t.Errorf("expected non-zero result time")
	}
	if repository.pricesAtTime.IsZero() {
		t.Errorf("expected non-zero time")
	}
	if !result.At.Equal(repository.pricesAtTime) {
		t.Errorf("expected result time %v, got %v", repository.pricesAtTime, result.At)
	}
}

// Проверяет, что GetPricesAt передает валидные данные в репозиторий
func TestGetPricesAtCallsRepository(t *testing.T) {
	repository := &fakePriceRepository{}
	service := NewPriceService(repository)
	at := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)

	result, err := service.GetPricesAt(context.Background(), []domain.GoodID{1, 2}, at)

	if err != nil {
		t.Errorf("get prices at: %v", err)
		return
	}
	if len(repository.pricesAtIDs) != 2 {
		t.Errorf("expected 2 good ids, got %d", len(repository.pricesAtIDs))
		return
	}
	if repository.pricesAtIDs[0] != 1 || repository.pricesAtIDs[1] != 2 {
		t.Errorf("unexpected good ids: %v", repository.pricesAtIDs)
	}
	if !repository.pricesAtTime.Equal(at) {
		t.Errorf("expected at %v, got %v", at, repository.pricesAtTime)
	}
	if !result.At.Equal(at) {
		t.Errorf("expected result at %v, got %v", at, result.At)
	}
}

// у периода истории должен быть from
func TestGetHistoryReturnsErrorForEmptyFrom(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.GetHistory(context.Background(), HistoryFilter{
		To: time.Date(2026, 6, 27, 11, 0, 0, 0, time.UTC),
	})

	if !errors.Is(err, ErrInvalidPeriod) {
		t.Errorf("expected ErrInvalidPeriod, got %v", err)
	}
}

// у периода истории должен быть to
func TestGetHistoryReturnsErrorForEmptyTo(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.GetHistory(context.Background(), HistoryFilter{
		From: time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC),
	})

	if !errors.Is(err, ErrInvalidPeriod) {
		t.Errorf("expected ErrInvalidPeriod, got %v", err)
	}
}

// Проверяет, что период истории должен быть упорядочен.
func TestGetHistoryReturnsErrorForInvalidPeriod(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})
	from := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)

	_, err := service.GetHistory(context.Background(), HistoryFilter{
		From: from,
		To:   from,
	})

	if !errors.Is(err, ErrInvalidPeriod) {
		t.Errorf("expected ErrInvalidPeriod, got %v", err)
	}
}

// Проверяет, что good_id истории должны быть положительными, если они переданы.
func TestGetHistoryReturnsErrorForInvalidGoodID(t *testing.T) {
	service := NewPriceService(&fakePriceRepository{})

	_, err := service.GetHistory(context.Background(), HistoryFilter{
		GoodIDs: []domain.GoodID{1, -1},
		From:    time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC),
		To:      time.Date(2026, 6, 27, 11, 0, 0, 0, time.UTC),
	})

	if !errors.Is(err, ErrInvalidGoodID) {
		t.Errorf("expected ErrInvalidGoodID, got %v", err)
	}
}

// Проверяет, что пустой список good_id означает историю по всем товарам.
func TestGetHistoryAllowsEmptyGoodIDs(t *testing.T) {
	repository := &fakePriceRepository{}
	service := NewPriceService(repository)
	from := time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 27, 11, 0, 0, 0, time.UTC)

	_, err := service.GetHistory(context.Background(), HistoryFilter{
		From: from,
		To:   to,
	})

	if err != nil {
		t.Errorf("get history: %v", err)
		return
	}
	if !repository.historyFilter.From.Equal(from) {
		t.Errorf("expected from %v, got %v", from, repository.historyFilter.From)
	}
	if !repository.historyFilter.To.Equal(to) {
		t.Errorf("expected to %v, got %v", to, repository.historyFilter.To)
	}
	if len(repository.historyFilter.GoodIDs) != 0 {
		t.Errorf("expected empty good ids, got %v", repository.historyFilter.GoodIDs)
	}
}
