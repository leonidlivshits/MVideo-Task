package http

import nethttp "net/http"

const (
	corsAllowMethods = "GET, POST, OPTIONS"
	corsAllowHeaders = "Accept, Authorization, Content-Type"
	corsExposeHeaders = "Content-Disposition"
	corsMaxAge = "600"
)

func NewRouter(handler *Handler) nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("POST /api/v1/prices", handler.SetPrices)
	mux.HandleFunc("GET /api/v1/prices", handler.GetPricesAt)
	mux.HandleFunc("GET /api/v1/prices/history.csv", handler.GetHistoryCSV)

	return withCORS(mux)
}

func withCORS(next nethttp.Handler) nethttp.Handler {
	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}

		w.Header().Add("Vary", "Origin")
		w.Header().Add("Vary", "Access-Control-Request-Method")
		w.Header().Add("Vary", "Access-Control-Request-Headers")
		w.Header().Set("Access-Control-Expose-Headers", corsExposeHeaders)

		if r.Method == nethttp.MethodOptions {
			allowHeaders := r.Header.Get("Access-Control-Request-Headers")
			if allowHeaders == "" {
				allowHeaders = corsAllowHeaders
			}

			w.Header().Set("Access-Control-Allow-Methods", corsAllowMethods)
			w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			w.Header().Set("Access-Control-Max-Age", corsMaxAge)
			w.WriteHeader(nethttp.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
