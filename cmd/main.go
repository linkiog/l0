package main

import (
	"context"
	"fmt"

	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/linkiog/lo/internal/cache"
	"github.com/linkiog/lo/internal/config"
	"github.com/linkiog/lo/internal/db"
	"github.com/linkiog/lo/internal/kafka"
	"github.com/linkiog/lo/internal/repository"
	"github.com/linkiog/lo/internal/server"
)

func main() {
	cfg := config.NewConfig()

	postgresDB, err := db.NewPostgresDB(cfg)
	if err != nil {
		panic(fmt.Errorf("failed to connect to Postgres: %w", err))
	}
	defer postgresDB.Close()

	repo := repository.NewRepository(postgresDB)

	c := cache.NewCache()

	if err := loadCacheFromDB(repo, c); err != nil {
		fmt.Println("Error loading orders into cache:", err)
	} else {
		fmt.Println("Cache loaded from DB successfully!")
	}

	ctx, cancel := context.WithCancel(context.Background())
	err = kafka.StartConsumerGroup(ctx, cfg, repo, c)
	if err != nil {
		panic(fmt.Errorf("failed to start consumer group: %w", err))
	}

	srv := &http.Server{
		Addr:    ":8080",
		Handler: server.NewServer(repo, c),
	}

	go func() {
		fmt.Println("Starting HTTP server on :8081")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Println("HTTP server error:", err)
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan
	fmt.Println("Shutting down gracefully...")

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Println("HTTP server Shutdown:", err)
	} else {
		fmt.Println("HTTP server stopped.")
	}

	fmt.Println("Service exited.")
}
func loadCacheFromDB(repo *repository.Repository, c *cache.Cache) error {
	orders, err := repo.GetAllOrders()
	if err != nil {
		return err
	}
	for _, o := range orders {
		c.Set(o)
	}
	return nil
}
