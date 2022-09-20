package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi"
)

const port = 8080

type httpServer struct {
	server *http.Server
}

func login(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("login"))
	if err != nil {
		log.Panicf("call to login failed: %#v", err)
	}
}

func verify(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write([]byte("verify"))
	if err != nil {
		log.Panicf("call to verify failed: %#v", err)
	}
}

func (s *httpServer) Start() error {
	router := chi.NewRouter()
	router.Post("/auth/api/v1/login", login)
	router.Post("/auth/api/v1/verify", verify)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	return s.server.ListenAndServe()
}

func (s *httpServer) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	server := httpServer{}
	go func() {
		if err := server.Start(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server failed: %#v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1 * time.Minute)
	defer cancel()

	err := server.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("server shutdown failed: %#v", err)
	}
}
