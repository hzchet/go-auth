package main

import (
	"os"
	"log"
	"errors"
	"net/http"
	"time"
	"context"
	"os/signal"
	"syscall"
	
)

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