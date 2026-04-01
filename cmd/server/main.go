package main

import (
	"context"
	"log"
	"mini-redis/internal/server"
	"mini-redis/internal/storage"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mini-redis/internal/config"
)

func main() {
	cfg := config.Default()
	store := storage.NewMemoryStorage()
	srv := server.New(cfg.Address, store)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	go func() {
		sig := <-sigCh
		log.Printf("received signal %s, shutting down", sig)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	log.Printf("mini-redis boot config: addr=%s aof=%s", cfg.Address, cfg.AOFPath)
	if err := srv.Start(); err != nil {
		log.Fatalf("server stopped with error: %v", err)
	}
}
