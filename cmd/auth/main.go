package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"server/internal/server"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	server := server.HttpServer{}
	go func() {
		if err := server.Start("config/config.yaml"); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server failed: %#v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err := server.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("server shutdown failed: %#v", err)
	}
}
