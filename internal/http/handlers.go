package http

import (
	nethttp "net/http"

	"mvideo-task/internal/service"
)

type Handler struct {
	service *service.PriceService
}

func NewHandler(service *service.PriceService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SetPrices(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeError(w, nethttp.StatusNotImplemented, "not implemented")
}

func (h *Handler) GetPricesAt(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeError(w, nethttp.StatusNotImplemented, "not implemented")
}

func (h *Handler) GetHistoryCSV(w nethttp.ResponseWriter, r *nethttp.Request) {
	writeError(w, nethttp.StatusNotImplemented, "not implemented")
}
