package main

import (
	"log/slog"
	"os"

	"github.com/hashicorp/go-cleanhttp"
	"github.com/senyosimpson/flywatch/flywatch"
	"go.etcd.io/bbolt"
)

func main() {
	logger := slog.Default()

	db, err := bbolt.Open("flywatch.db", 0600, nil)
	if err != nil {
		logger.Error("error opening database", "error", err)
		os.Exit(-1)
	}
	defer db.Close()

	flyAPIToken := os.Getenv("FLY_API_TOKEN")
	if flyAPIToken == "" {
		logger.Error("FLY_API_TOKEN is not set")
		os.Exit(1)
	}

	httpClient := cleanhttp.DefaultClient()
	controller := flywatch.Controller{
		Db:       db,
		Logger:   logger,
		Client:   httpClient,
		APIToken: flyAPIToken,
	}
	go controller.Run()

	fw := flywatch.Flywatch{
		Logger: logger,
		Db:     db,
	}
	fw.Run()
}
