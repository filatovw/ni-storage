package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/filatovw/ni-storage/api"
	"github.com/filatovw/ni-storage/config"
	"github.com/filatovw/ni-storage/engine/narwal"
)

func main() {
	config := config.Load()
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %s", err)
	}
	sugarLogger := logger.Sugar()

	storage, err := narwal.New(config.NarWAL.DataDir, sugarLogger)
	if err != nil {
		log.Printf("failed to init storage: %s", err)
		return
	}

	server := api.New(sugarLogger, storage, config)
	if err := server.ListenAndServe(); err != nil {
		log.Printf("stopped with error: %s", err)
	}
}
