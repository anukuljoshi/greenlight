package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/anukuljoshi/greenlight/internal/data"
	"github.com/anukuljoshi/greenlight/internal/jsonlog"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

const version = "1.0.0"

// config struct to hold settings for our application
type config struct {
	port int
	env string
	db struct {
		dsn string
		maxOpenConns int
		maxIdleConns int
		maxIdleTime string
	}
	limiter struct {
		rps float64
		burst int
		enabled bool
	}
}

// application struct to hold dependencies for handlers, middlewares, helpers
type application struct {
	config config
	logger *jsonlog.Logger
	models data.Models
}

func main() {
	var cfg config

	// read flag vars
	flag.IntVar(&cfg.port, "port", 4000, "PORT for application")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")

	// read db connection pool settings from flags
	flag.IntVar(&cfg.db.maxOpenConns, "db-max-open-conns", 25, "PostgreSQL max open connections")
	flag.IntVar(&cfg.db.maxIdleConns, "db-max-idle-conns", 25, "PostgreSQL max idle connections")
	flag.StringVar(&cfg.db.maxIdleTime, "db-max-idle-time", "15m", "PostgreSQL max idle connection time")
	// read limiter settings from flag vars
	flag.Float64Var(&cfg.limiter.rps, "limiter-rps", 2, "Rate limiter maximum requests per second")
	flag.IntVar(&cfg.limiter.burst, "limiter-burst", 4, "Rate limiter maximum burst")
	flag.BoolVar(&cfg.limiter.enabled, "limiter-enabled", true, "Enable rate limiter")
	flag.Parse()

	// create logger
	var logger = jsonlog.New(os.Stdout, jsonlog.LevelInfo)

	// read env
	err := godotenv.Load(".env")
	if err!=nil {
		logger.PrintFatal(err, nil)
	}
	cfg.db.dsn = os.Getenv("DSN")


	// connect to db
	db, err := openDB(cfg)
	if err!=nil {
		logger.PrintFatal(err, nil)
	}
	defer db.Close()

	logger.PrintInfo("database connected", nil)

	// create app struct
	var app = &application{
		config: cfg,
		logger: logger,
		models: data.NewModels(db),
	}

	// create server with config
	var srv = &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		// built in error log will use out custom logger for logging
		ErrorLog: log.New(logger, "", 0),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": cfg.env,
	})
	err = srv.ListenAndServe()
	logger.PrintFatal(err, nil)
}

func openDB(cfg config) (*sql.DB, error) {
	db, err := sql.Open("postgres", cfg.db.dsn)
	if err!=nil {
		return nil, err
	}
	// set db connection pool settings
	db.SetMaxOpenConns(cfg.db.maxOpenConns)
	db.SetMaxIdleConns(cfg.db.maxIdleConns)
	duration, err := time.ParseDuration(cfg.db.maxIdleTime)
	if err!=nil {
		return nil, err
	}
	db.SetConnMaxIdleTime(duration)

	// create context with 5 second deadline
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	// establish new connection with context, returns error if not connected in 5 seconds
	err = db.PingContext(ctx)
	if err!=nil {
		return nil, err
	}
	return db, nil
}
