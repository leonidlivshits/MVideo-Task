package http

import nethttp "net/http"

func writeError(w nethttp.ResponseWriter, status int, message string) {
	nethttp.Error(w, message, status)
}
