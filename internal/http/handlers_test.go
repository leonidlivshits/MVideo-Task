package http

import (
	"context"
	"encoding/json"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"mvideo-task/internal/domain"
	"mvideo-task/internal/service"
)

type fakePriceRepository struct {
	insertPrices func(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error)
	getPricesAt  func(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error)
	getHistory   func(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error)
}

func (r *fakePriceRepository) InsertPrices(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
	if r.insertPrices == nil {
		return nil, errors.New("unexpected InsertPrices call")
	}

	return r.insertPrices(ctx, prices)
}

func (r *fakePriceRepository) GetPricesAt(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
	if r.getPricesAt == nil {
		return nil, errors.New("unexpected GetPricesAt call")
	}

	return r.getPricesAt(ctx, goodIDs, at)
}

func (r *fakePriceRepository) GetHistory(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error) {
	if r.getHistory == nil {
		return nil, errors.New("unexpected GetHistory call")
	}

	return r.getHistory(ctx, filter)
}

var testTime = time.Date(2026, 6, 27, 10, 0, 0, 0, time.UTC)

// создает цены и возвращает созданные записи
func TestSetPricesHandlerCreatesPrices(t *testing.T) {
	var insertedPrices []domain.PricePoint
	repository := &fakePriceRepository{
		insertPrices: func(ctx context.Context, prices []domain.PricePoint) ([]domain.PricePoint, error) {
			insertedPrices = prices

			result := make([]domain.PricePoint, 0, len(prices))
			for _, price := range prices {
				price.CreateAt = testTime
				result = append(result, price)
			}

			return result, nil
		},
	}
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
	if len(insertedPrices) != 1 {
		t.Errorf("expected 1 inserted price, got %d", len(insertedPrices))
		return
	}
	if insertedPrices[0].GoodID != 1 || insertedPrices[0].Price != 100 {
		t.Errorf("unexpected inserted price: %+v", insertedPrices[0])
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
	var pricesAtIDs []domain.GoodID
	var pricesAtTime time.Time
	repository := &fakePriceRepository{
		getPricesAt: func(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
			pricesAtIDs = goodIDs
			pricesAtTime = at

			return []domain.PriceAt{
				{GoodID: 1, Price: &price, CreateAt: &testTime},
			}, nil
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
	if len(pricesAtIDs) != 2 {
		t.Errorf("expected 2 good ids, got %d", len(pricesAtIDs))
		return
	}
	if pricesAtIDs[0] != 1 || pricesAtIDs[1] != 2 {
		t.Errorf("unexpected good ids: %v", pricesAtIDs)
	}
	if !pricesAtTime.Equal(testTime) {
		t.Errorf("expected at %v, got %v", testTime, pricesAtTime)
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
	var pricesAtTime time.Time
	repository := &fakePriceRepository{
		getPricesAt: func(ctx context.Context, goodIDs []domain.GoodID, at time.Time) ([]domain.PriceAt, error) {
			pricesAtTime = at

			return []domain.PriceAt{}, nil
		},
	}
	router := newTestRouter(repository)

	request := httptest.NewRequest(nethttp.MethodGet, "/api/v1/prices?good_id=1", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Errorf("expected status %d, got %d", nethttp.StatusOK, response.Code)
	}
	if pricesAtTime.IsZero() {
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

// GET /api/v1/prices/history.csv возвращает историю цен в CSV
func TestGetHistoryCSVHandlerReturnsCSV(t *testing.T) {
	secondTime := testTime.Add(time.Hour)
	var historyFilter service.HistoryFilter
	repository := &fakePriceRepository{
		getHistory: func(ctx context.Context, filter service.HistoryFilter) ([]domain.PricePoint, error) {
			historyFilter = filter

			return []domain.PricePoint{
				{GoodID: 1, CreateAt: testTime, Price: 100},
				{GoodID: 2, CreateAt: secondTime, Price: 200},
			}, nil
		},
	}
	router := newTestRouter(repository)

	request := httptest.NewRequest(
		nethttp.MethodGet,
		"/api/v1/prices/history.csv?good_id=1,2&from=2026-06-27T10:00:00Z&to=2026-06-27T12:00:00Z",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusOK {
		t.Errorf("expected status %d, got %d", nethttp.StatusOK, response.Code)
	}
	if response.Header().Get("Content-Type") != "text/csv" {
		t.Errorf("expected content type text/csv, got %q", response.Header().Get("Content-Type"))
	}
	if len(historyFilter.GoodIDs) != 2 {
		t.Errorf("expected 2 good ids, got %d", len(historyFilter.GoodIDs))
		return
	}
	if historyFilter.GoodIDs[0] != 1 || historyFilter.GoodIDs[1] != 2 {
		t.Errorf("unexpected good ids: %v", historyFilter.GoodIDs)
	}
	if !historyFilter.From.Equal(testTime) {
		t.Errorf("expected from %v, got %v", testTime, historyFilter.From)
	}
	if !historyFilter.To.Equal(testTime.Add(2 * time.Hour)) {
		t.Errorf("expected to %v, got %v", testTime.Add(2*time.Hour), historyFilter.To)
	}

	expectedBody := "good_id,create_at,price\n" +
		"1,2026-06-27T10:00:00Z,100\n" +
		"2,2026-06-27T11:00:00Z,200\n"
	if response.Body.String() != expectedBody {
		t.Errorf("unexpected body:\n%s", response.Body.String())
	}
}

// GET /api/v1/prices/history.csv требует from
func TestGetHistoryCSVHandlerReturnsErrorForMissingFrom(t *testing.T) {
	router := newTestRouter(&fakePriceRepository{})

	request := httptest.NewRequest(
		nethttp.MethodGet,
		"/api/v1/prices/history.csv?to=2026-06-27T12:00:00Z",
		nil,
	)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != nethttp.StatusBadRequest {
		t.Errorf("expected status %d, got %d", nethttp.StatusBadRequest, response.Code)
	}
}

func newTestRouter(repository service.PriceRepository) nethttp.Handler {
	priceService := service.NewPriceService(repository)
	handler := NewHandler(priceService)

	return NewRouter(handler)
}
