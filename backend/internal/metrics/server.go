package metrics

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Server serves /metrics and /debug/pprof on a dedicated listen address.
type Server struct {
	httpServer *http.Server
	addr       string
}

// NewServer builds a metrics/debug HTTP server. Empty authToken allows unauthenticated access (development).
func NewServer(addr, authToken string) *Server {
	mux := http.NewServeMux()
	metricsHandler := promhttp.HandlerFor(Registry, promhttp.HandlerOpts{Registry: Registry})
	mux.Handle("/metrics", withAuth(authToken, metricsHandler))
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok","service":"plexus-metrics"}`))
	})

	mux.HandleFunc("/debug/pprof/", withAuthFunc(authToken, pprof.Index))
	mux.HandleFunc("/debug/pprof/cmdline", withAuthFunc(authToken, pprof.Cmdline))
	mux.HandleFunc("/debug/pprof/profile", withAuthFunc(authToken, pprof.Profile))
	mux.HandleFunc("/debug/pprof/symbol", withAuthFunc(authToken, pprof.Symbol))
	mux.HandleFunc("/debug/pprof/trace", withAuthFunc(authToken, pprof.Trace))
	mux.Handle("/debug/pprof/heap", withAuth(authToken, pprof.Handler("heap")))
	mux.Handle("/debug/pprof/goroutine", withAuth(authToken, pprof.Handler("goroutine")))
	mux.Handle("/debug/pprof/threadcreate", withAuth(authToken, pprof.Handler("threadcreate")))
	mux.Handle("/debug/pprof/block", withAuth(authToken, pprof.Handler("block")))
	mux.Handle("/debug/pprof/mutex", withAuth(authToken, pprof.Handler("mutex")))
	mux.Handle("/debug/pprof/allocs", withAuth(authToken, pprof.Handler("allocs")))

	return &Server{
		addr: addr,
		httpServer: &http.Server{
			Addr:              addr,
			Handler:           mux,
			ReadHeaderTimeout: 5 * time.Second,
		},
	}
}

func (s *Server) Addr() string { return s.addr }

// Start listens until the server is shut down. Blocks.
func (s *Server) Start() error {
	slog.Info("metrics server listening", "addr", s.addr)
	err := s.httpServer.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}

// Shutdown gracefully stops the metrics server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func withAuth(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !authorized(token, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withAuthFunc(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !authorized(token, r) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)
	}
}

func authorized(token string, r *http.Request) bool {
	if token == "" {
		return true
	}
	auth := r.Header.Get("Authorization")
	if strings.HasPrefix(auth, "Bearer ") && strings.TrimPrefix(auth, "Bearer ") == token {
		return true
	}
	return r.Header.Get("X-Metrics-Token") == token
}
