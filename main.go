package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/krishnassh/picostatus/internal/config"
	"github.com/krishnassh/picostatus/internal/scheduler"
	"github.com/krishnassh/picostatus/internal/server"
	"github.com/krishnassh/picostatus/internal/storage"
)

func main() {
	cfgPath := "config.toml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db, err := storage.Open("picostatus.db")
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}

	repo := storage.NewRepository(db)

	if err := repo.SyncChecks(cfg.Checks); err != nil {
		log.Fatalf("failed to sync checks: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sched := scheduler.New(repo)
	go sched.Start(ctx)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.New(repo),
	}

	go func() {
		log.Println("listening on http://localhost:8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
}
