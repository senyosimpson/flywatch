package main

import (
	"log/slog"
	"os"

	"github.com/senyosimpson/flywatch/flywatch"
	"go.etcd.io/bbolt"
)

func main() {
	// logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	logger := slog.Default()

	db, err := bbolt.Open("flywatch.db", 0600, nil)
	if err != nil {
		logger.Error("error opening database", "error", err)
		os.Exit(-1)
	}
	defer db.Close()

	// controller := flywatch.Controller{
	// 	Db:     db,
	// 	Logger: logger,
	// }
	// go controller.Run()

	fw := flywatch.Flywatch{
		Logger: logger,
		Db:     db,
	}
	fw.Run()
}
