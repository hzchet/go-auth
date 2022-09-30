package server

import (
	"server/internal/pkg/controller"

	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

type HttpServer struct {
	server *http.Server
}

func (s *HttpServer) Start() error {
	cntrl := controller.New("config/config.yaml")

	router := chi.NewRouter()
	router.Post(cntrl.Config.Endpoints["login"], cntrl.Login)
	router.Post(cntrl.Config.Endpoints["verify"], cntrl.Verify)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cntrl.Config.Port),
		Handler: router,
	}

	return s.server.ListenAndServe()
}

func (s *HttpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
