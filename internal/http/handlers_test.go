package http

import (
	"context"
	"encoding/json"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mvideo-task/internal/domain"
	"mvideo-task/internal/service"
)

type fakePriceRepository struct {
	insertedPrices []domain.PricePoint
	pricesAtIDs    []domain.GoodID
	pricesAtTime   time.Time
	pricesAtResult []domain.PriceAt
}

func (r *fakePriceRepository) InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	r.insertedPrices = prices

	result := make([]domain.PricePoint, 0, len(prices))
	for _, price := range prices {
		price.CreateAt = testTime
		result = append(result, price)
	}

	return result, nil
}

func (r *fakePriceRepository) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	r.pricesAtIDs = goodIDs
	r.pricesAtTime = at

	return r.pricesAtResult, nil
}

func (r *fakePriceRepository) GetHistory(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error) {
	return []domain.PricePoint{}, nil
}

var testTime = time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)

// создает цены и возвращает созданные записи
func TestSetPricesHandlerCreatesPrices(t *testing.T) {
	repository := &fakePriceRepository{}
	router := newTestRouter(repository)

	request := httptest.NewRequest(
		nethttp.MethodPost,
		"/api/v1/prices",
		strings.NewReader(`{"items":[{"good_id":1,"price":100}]}`),
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusCreated {
		t.Errorf("expected status %d, got %d", nethttp.StatusCreated, response.Code)
	}
	if len(repository.insertedPrices) != 1 {
		t.Errorf("expected 1 inserted price, got %d", len(repository.insertedPrices))
		return
	}
	if repository.insertedPrices[0].GoodID != 1 || repository.insertedPrices[0].Price != 100 {
		t.Errorf("unexpected inserted price: %+v", repository.insertedPrices[0])
	}

	var body SetPricesResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Errorf("decode response: %v", err)
		return
	}
	if len(body.Items) != 1 {
		t.Errorf("expected 1 response item, got %d", len(body.Items))
		return
	}
	if body.Items[0].GoodID != 1 || body.Items[0].Price != 100 {
		t.Errorf("unexpected response item: %+v", body.Items[0])
	}
	if !body.Items[0].CreateAt.Equal(testTime) {
		t.Errorf("expected create_at %v, got %v", testTime, body.Items[0].CreateAt)
	}
}

// отклоняет неизвестные поля в JSON
func TestSetPricesHandlerRejectsUnknownJSONField(t *testing.T) {
	router := newTestRouter(&fakePriceRepository{})

	request := httptest.NewRequest(
		nethttp.MethodPost,
		"/api/v1/prices",
		strings.NewReader(`{"items":[{"good_id":1,"price":100,"currency":"RUB"}]}`),
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", nethttp.StatusBadRequest, response.Code)
	}
}

// возвращает цены на переданный момент времени
func TestGetPricesAtHandlerReturnsPrices(t *testing.T) {
	price := domain.Price(100)
	repository := &fakePriceRepository{
		pricesAtResult: []domain.PriceAt{
			{GoodID: 1, Price: &price, CreateAt: &testTime},
		},
	}
	router := newTestRouter(repository)

	request := httptest.NewRequest(
		nethttp.MethodGet,
		"/api/v1/prices?good_id=1,2&at=2026-06-27T10:00:00Z",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Errorf("expected status %d, got %d", nethttp.StatusOK, response.Code)
	}
	if len(repository.pricesAtIDs) != 2 {
		t.Errorf("expected 2 good ids, got %d", len(repository.pricesAtIDs))
		return
	}
	if repository.pricesAtIDs[0] != 1 || repository.pricesAtIDs[1] != 2 {
		t.Errorf("unexpected good ids: %v", repository.pricesAtIDs)
	}
	if !repository.pricesAtTime.Equal(testTime) {
		t.Errorf("expected at %v, got %v", testTime, repository.pricesAtTime)
	}

	var body PricesAtResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Errorf("decode response: %v", err)
		return
	}
	if !body.At.Equal(testTime) {
		t.Errorf("expected response at %v, got %v", testTime, body.At)
	}
	if len(body.Items) != 1 {
		t.Errorf("expected 1 response item, got %d", len(body.Items))
		return
	}
	if body.Items[0].GoodID != 1 || body.Items[0].Price == nil || *body.Items[0].Price != 100 {
		t.Errorf("unexpected response item: %+v", body.Items[0])
	}
}

// использует текущее время, если at не передали
func TestGetPricesAtHandlerUsesCurrentTimeWhenAtIsMissing(t *testing.T) {
	repository := &fakePriceRepository{}
	router := newTestRouter(repository)

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/prices?good_id=1", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Errorf("expected status %d, got %d", nethttp.StatusOK, response.Code)
	}
	if repository.pricesAtTime.IsZero() {
		t.Errorf("expected non-zero repository time")
	}

	var body PricesAtResponse
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Errorf("decode response: %v", err)
		return
	}
	if body.At.IsZero() {
		t.Errorf("expected non-zero response at")
	}
}

// требует хотя бы один good_id
func TestGetPricesAtHandlerReturnsErrorForMissingGoodID(t *testing.T) {
	router := newTestRouter(&fakePriceRepository{})

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/prices", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", nethttp.StatusBadRequest, response.Code)
	}
}

func newTestRouter(repository *fakePriceRepository) nethttp.Handler {
	priceService := service.NewPriceService(repository)
	handler := NewHandler(priceService)

	return NewRouter(handler)
}
