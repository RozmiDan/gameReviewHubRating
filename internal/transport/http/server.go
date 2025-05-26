package http_serv

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func New(logger *zap.Logger) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	return mux
}
