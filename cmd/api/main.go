package main

import (
	"flag"
	"fmt"
	"github.com/dragonsecurity/breachdetector/internal/database"
	"github.com/dragonsecurity/breachdetector/internal/version"
	"github.com/lmittmann/tint"
	"log/slog"
	"os"
	"runtime/debug"
	"sync"
)

type config struct {
	baseURL  string
	httpPort int
	db       struct {
		dsn string
	}
}

type application struct {
	config config
	db     *database.DB
	logger *slog.Logger
	wg     sync.WaitGroup
}

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))
	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:9999", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 9999, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "postgres:password@localhost:5432/breachdetector?sslmode=disable", "postgreSQL DSN")

	showVersion := flag.Bool("version", false, "display version and exit")

	flag.Parse()

	if *showVersion {
		fmt.Printf("version: %s\n", version.Get())
		return nil
	}

	db, err := database.New(cfg.db.dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	app := &application{
		config: cfg,
		db:     db,
		logger: logger,
	}

	return app.serveHTTP()
}
