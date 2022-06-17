package main

import (
	"net/http"
	"os"

	"github.com/athosone/golib/examples/media-type-versioning/books"
	glogger "github.com/athosone/golib/pkg/logger"
	"github.com/athosone/golib/pkg/server"
	gmiddleware "github.com/athosone/golib/pkg/server/middleware"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

var (
	logger  *zap.SugaredLogger
	Version string = "dev"
)

// Will launch two servers based on two libraries:
// - Gorilla Mux: http://localhost:8080/api/books
// - Go-Chi:      http://localhost:8081/api/books
func main() {
	// Init logger
	logger = glogger.NewLogger(os.Getenv("IS_DEBUG") == "true").With("service", "sample-service").With("version", Version)
	defer logger.Sync()

	zap.S().Info("Starting sample service version: ", Version)
	// Setup gorilla mux server

	routerMux := mux.NewRouter()
	routerMux.Use(gmiddleware.CompressResponse())
	routerMux.Use(gmiddleware.InjectLoggerInRequest(func(r *http.Request) *zap.SugaredLogger {
		return logger.With("request_id", middleware.GetReqID(r.Context())).With("router", "mux")
	}))
	routerMux.Use(gmiddleware.RequestLogger([]string{"/healthy", "/ready"}))
	routerMux.Use(middleware.Heartbeat("/healthy"))

	setupWithMux(routerMux)

	// Setup chi server
	routerChi := chi.NewRouter()

	routerChi.Use(gmiddleware.CompressResponse())
	routerChi.Use(gmiddleware.InjectLoggerInRequest(func(r *http.Request) *zap.SugaredLogger {
		return logger.With("request_id", middleware.GetReqID(r.Context())).With("router", "chi")
	}))
	routerChi.Use(gmiddleware.RequestLogger([]string{"/healthy", "/ready"}))
	routerChi.Use(middleware.Heartbeat("/healthy"))

	setupWithChi(routerChi)

	// Config server
	srvMux := &http.Server{Addr: "0.0.0.0:8080", Handler: routerMux}
	zap.S().Infow("Starting mux server", "addr", srvMux.Addr)
	go server.ListenAndServe(srvMux)

	srvChi := &http.Server{Addr: "0.0.0.0:8081", Handler: routerChi}
	zap.S().Infow("Starting chi server", "addr", srvChi.Addr)
	server.ListenAndServe(srvChi)
}

func setupWithMux(router *mux.Router) {
	router.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}).Methods("GET")
	apiRouter := router.PathPrefix("/api").Subrouter()

	books.SetupMux(apiRouter)
}

func setupWithChi(router chi.Router) {
	router.Get("/ready", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	router.Route("/api", func(r chi.Router) {
		books.SetupWithChi(r)
	})
}
