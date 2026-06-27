package http

import (
	"encoding/json"
	"errors"
	nethttp "net/http"

	"mvideo-task/internal/service"
)

const (
	errorMessageInvalidJSON = "invalid json"
	errorMessageItemsRequired = "items are required"
	errorMessageGoodIDRequired = "good_id is required"
	errorMessageInvalidGoodID = "invalid good_id %q"
	errorMessageGoodIDMustPositive = "good_id must be positive"
	errorMessagePriceMustPositive = "price must be positive"
	errorMessageInvalidPeriod = "invalid period"
	errorMessageAtRequired = "at is required"
	errorMessageAtMustProvidedOnce = "at must be provided once"
	errorMessageInvalidAt = "invalid at, use RFC3339 format"
	errorMessageInternalServerError = "internal server error"
)

func writeError(w nethttp.ResponseWriter, status int, message string) {
	writeJSON(w, status, ErrorResponse{Error: message})
}

func writeServiceError(w nethttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrEmptyPrices):
		writeError(w, nethttp.StatusBadRequest, errorMessageItemsRequired)
	case errors.Is(err, service.ErrEmptyGoodIDs):
		writeError(w, nethttp.StatusBadRequest, errorMessageGoodIDRequired)
	case errors.Is(err, service.ErrInvalidGoodID):
		writeError(w, nethttp.StatusBadRequest, errorMessageGoodIDMustPositive)
	case errors.Is(err, service.ErrInvalidPrice):
		writeError(w, nethttp.StatusBadRequest, errorMessagePriceMustPositive)
	case errors.Is(err, service.ErrInvalidPeriod):
		writeError(w, nethttp.StatusBadRequest, errorMessageInvalidPeriod)
	default:
		writeError(w, nethttp.StatusInternalServerError, errorMessageInternalServerError)
	}
}

func writeJSON(w nethttp.ResponseWriter, status int, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(response)
}
