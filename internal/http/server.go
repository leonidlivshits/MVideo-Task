package http

import (
	nethttp "net/http"
	"time"

	"mvideo-task/internal/service"
)

const readHeaderTimeout = 5 * time.Second

func NewServer(addr string, priceService *service.PriceService) *nethttp.Server {
	handler := NewHandler(priceService)

	return &nethttp.Server{
		Addr:              addr,
		Handler:           NewRouter(handler),
		ReadHeaderTimeout: readHeaderTimeout,
	}
}
