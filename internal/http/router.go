package http

import nethttp "net/http"

func NewRouter(handler *Handler) nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("POST /api/v1/prices", handler.SetPrices)
	mux.HandleFunc("GET /api/v1/prices", handler.GetPricesAt)
	mux.HandleFunc("GET /api/v1/prices/history.csv", handler.GetHistoryCSV)

	return mux
}
