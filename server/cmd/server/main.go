package main

import (
	"belaz-calendar-server/internal/pkg/database"
	"belaz-calendar-server/internal/pkg/logger"
	"belaz-calendar-server/internal/server"
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	logger.Init()

	db := database.ConnectDB()

	port := os.Getenv("PORT")
	if port == "" {
		port = ":3030"
	}

	if err := db.Ping(); err != nil {
		logger.Log.Fatal("Database ping failed", zap.Error(err))
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	logger.Log.Info("Database connected successfully")

	srv := server.NewServer(db, logger.Log)

	go func() {
		logger.Log.Info("Starting server", zap.String("port", port))
		if err := srv.Run(port); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	flag := os.Getenv("SKIP_MAINTENANCE_INIT") == "true"

	if !flag {
		logger.Log.Info("Initializing maintenance schedules...")
		if err := srv.InitializeMaintenanceSchedules(context.Background()); err != nil {
			logger.Log.Error("Failed to initialize maintenance schedules", zap.Error(err))
		}
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Log.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Log.Error("Server forced to shutdown", zap.Error(err))
	}

	if err := db.Close(); err != nil {
		logger.Log.Error("Database connection close failed", zap.Error(err))
	}

	logger.Log.Info("Server exited gracefully")
}
