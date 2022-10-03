package server

import (
	"server/internal/pkg/controller"
	"server/internal/pkg/metrics"

	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/juju/zaputil/zapctx"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/riandyrn/otelchi"
	"go.uber.org/zap"
	"moul.io/chizap"
)

type HttpServer struct {
	server *http.Server
}

func (s *HttpServer) Start() error {
	logger, err := metrics.GetLogger(false, metrics.DSN, "myenv")
	if err != nil {
		log.Fatal(err)
	}
	
	logger.Info("Start!")

	if err := metrics.InitOtel(); err != nil {
		logger.Fatal("OTEL init", zap.Error(err))
	}

	ctx := zapctx.WithLogger(context.Background(), logger)

	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(otelchi.Middleware(metrics.TRACER_NAME, otelchi.WithChiRoutes(router)))
	router.Use(chizap.New(logger, &chizap.Opts{
		WithReferer:   true,
		WithUserAgent: true,
	}))

	cntrl := controller.New("config/config.yaml", ctx)
	router.Post(cntrl.Config.Endpoints["login"], cntrl.Login)
	router.Post(cntrl.Config.Endpoints["verify"], cntrl.Verify)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cntrl.Config.Port),
		Handler: router,
	}

	http.Handle("/metrics", promhttp.Handler())
	go http.ListenAndServe(":9000", nil)
	
	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
