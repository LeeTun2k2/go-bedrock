package serverx

import (
	"net/http"
	"net/http/pprof"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// newAdminMux builds the admin router: status probe, Prometheus metrics, and
// Go profiling endpoints.
func newAdminMux(s *Server) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/status", s.statusHandler)
	mux.Handle("/metrics", promhttp.Handler())

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	return mux
}
