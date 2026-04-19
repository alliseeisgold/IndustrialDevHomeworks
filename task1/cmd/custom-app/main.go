package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"industrialdev/task1/internal/app"
	"industrialdev/task1/internal/config"
)

func main() {
	cfg := config.Load()

	logger := log.New(
		os.Stdout,
		fmt.Sprintf("[%s] ", strings.ToUpper(cfg.LogLevel)),
		log.LstdFlags|log.Lmsgprefix,
	)

	api, err := app.New(cfg, logger)
	if err != nil {
		logger.Fatalf("create app: %v", err)
	}

	logger.Printf("custom app is listening on port %s", cfg.Port)

	if err := http.ListenAndServe(":"+cfg.Port, api.Handler()); err != nil {
		logger.Fatalf("server failed: %v", err)
	}
}
