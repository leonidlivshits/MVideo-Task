package http

import (
	"encoding/json"
	"errors"
	"fmt"
	nethttp "net/http"
	"strconv"
	"strings"
	"time"

	"mvideo-task/internal/domain"
	"mvideo-task/internal/service"
)

type Handler struct {
	service *service.PriceService
}

func NewHandler(service *service.PriceService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetPrices(w nethttp.ResponseWriter, r *nethttp.Request) {
	var request SetPricesRequest

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&request); err != nil {
		writeError(w, nethttp.StatusBadRequest, errorMessageInvalidJSON)
		return
	}

	prices := make([]domain.PricePoint, 0, len(request.Items))
	for _, item := range request.Items {
		prices = append(prices, domain.PricePoint{
			GoodID: domain.GoodID(item.GoodID),
			Price:  domain.Price(item.Price),
		})
	}

	createdPrices, err := h.service.SetPrices(r.Context(), prices)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := SetPricesResponse{
		Items: make([]SetPriceResponse, 0, len(createdPrices)),
	}
	for _, price := range createdPrices {
		response.Items = append(response.Items, SetPriceResponse{
			GoodID:   int64(price.GoodID),
			Price:    int(price.Price),
			CreateAt: price.CreateAt,
		})
	}

	writeJSON(w, nethttp.StatusCreated, response)
}

func (h *Handler) GetPricesAt(w nethttp.ResponseWriter, r *nethttp.Request) {
	query := r.URL.Query()

	goodIDs, err := parseGoodIDs(query["good_id"])
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	at, err := parseAt(query["at"])
	if err != nil {
		writeError(w, nethttp.StatusBadRequest, err.Error())
		return
	}

	result, err := h.service.GetPricesAt(r.Context(), goodIDs, at)
	if err != nil {
		writeServiceError(w, err)
		return
	}

	response := PricesAtResponse{
		At:    result.At,
		Items: make([]PriceResponse, 0, len(result.Items)),
	}
	for _, price := range result.Items {
		response.Items = append(response.Items, priceAtResponse(price))
	}

	writeJSON(w, nethttp.StatusOK, response)
}

func (h *Handler) GetHistoryCSV(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeError(w, nethttp.StatusNotImplemented, "not implemented")
}

func parseGoodIDs(values []string) ([]domain.GoodID, error) {
	goodIDs := make([]domain.GoodID, 0, len(values))

	for _, value := range values {
		parts := strings.Split(value, ",")
		for _, part := range parts {
			goodID, err := parseGoodID(part)
			if err != nil {
				return nil, err
			}

			goodIDs = append(goodIDs, goodID)
		}
	}

	return goodIDs, nil
}

func parseGoodID(value string) (domain.GoodID, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, errors.New(errorMessageGoodIDRequired)
	}

	goodID, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf(errorMessageInvalidGoodID, value)
	}

	return domain.GoodID(goodID), nil
}

func parseAt(values []string) (time.Time, error) {
	if len(values) == 0 {
		return time.Time{}, nil
	}

	if len(values) > 1 {
		return time.Time{}, errors.New(errorMessageAtMustProvidedOnce)
	}

	value := strings.TrimSpace(values[0])
	if value == "" {
		return time.Time{}, errors.New(errorMessageAtRequired)
	}

	at, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}, errors.New(errorMessageInvalidAt)
	}

	return at, nil
}

func priceAtResponse(price domain.PriceAt) PriceResponse {
	var responsePrice *int
	if price.Price != nil {
		value := int(*price.Price)
		responsePrice = &value
	}

	var createAt *time.Time
	if price.CreateAt != nil {
		value := *price.CreateAt
		createAt = &value
	}

	return PriceResponse{
		GoodID:   int64(price.GoodID),
		Price:    responsePrice,
		CreateAt: createAt,
	}
}
