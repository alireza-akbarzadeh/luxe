package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alireza-akbarzadeh/luxe/internal/config"
	"github.com/alireza-akbarzadeh/luxe/internal/shared/utils"
	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 15 * time.Second

// bootStrap runs the HTTP server and shuts down gracefully on SIGINT/SIGTERM.
func bootStrap(engine *gin.Engine, cfg *config.Config, onShutdown func()) {
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      engine,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		utils.Log.Infof("starting server on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			utils.Log.WithError(err).Fatal("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit
	utils.Log.WithField("signal", sig.String()).Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		utils.Log.WithError(err).Error("HTTP server shutdown error")
	} else {
		utils.Log.Info("HTTP server stopped")
	}

	if onShutdown != nil {
		onShutdown()
	}

	utils.Log.Info("shutdown complete")
}
