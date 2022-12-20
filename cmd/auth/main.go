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
	"server/internal/database/dbuser"
	dbUserMigrations "server/migrations/dbuser"
)

func main() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	server := server.HttpServer{}

	dbUser, err := dbuser.NewDatabase(ctx, "postgres://postgres:postgres@postgres:5432/postgres?sslmode=disable", dbUserMigrations.MigrationAssets)
	if err != nil {
		log.Fatalf("cannot create db: %v", err)
	}
	defer dbUser.Close(ctx)

	err = dbUser.Ping(ctx)
	if err != nil {
		log.Fatalf("cannot ping db: %v", err)
		return
	}
	go func() {
		if err := server.Start("config/config.yaml", dbUser); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("start server failed: %#v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 1*time.Minute)
	defer cancel()

	err = server.Stop(shutdownCtx)
	if err != nil {
		log.Fatalf("server shutdown failed: %#v", err)
	}
}
