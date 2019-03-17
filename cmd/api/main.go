package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"

	"github.com/filatovw/ni-storage/api"
	"github.com/filatovw/ni-storage/config"
	"github.com/filatovw/ni-storage/engine/narwal"
)

func main() {
	config := config.Load()

	zapConfig := zap.NewProductionConfig()
	if config.Debug {
		zapConfig.Level.SetLevel(zap.DebugLevel)
	}
	logger, err := zapConfig.Build()
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}
	slog := logger.Sugar()

	if config.Debug {
		slog.Infof("config %#v", config)
	}

	// try to shutdown application gracefully on SIGINT|SIGTERM
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	// init storage
	storage, err := narwal.New(ctx, config.NarWAL.DataDir, slog)
	if err != nil {
		log.Printf("failed to init storage: %s", err)
		return
	}

	server := api.New(ctx, slog, storage, *config)

	go func() {
		sig := <-sigs
		slog.Infof("Stopped with signal: %v", sig)

		// stop server gracefully
		if err := server.Shutdown(ctx); err != nil {
			slog.Errorf("server shutdowned with error: %s", err)
		}

		// stop storage goroutines
		cancel()
	}()

	// start web server
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		slog.Errorf("server stopped with error: %s", err)
	}
}
