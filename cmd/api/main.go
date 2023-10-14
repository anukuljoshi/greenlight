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
}

// application struct to hold dependencies for handlers, middlewares, helpers
type application struct {
	config config
	logger *log.Logger
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
	flag.Parse()

	// read env
	err := godotenv.Load(".env")
	if err!=nil {
		log.Fatal(err)
	}
	cfg.db.dsn = os.Getenv("DSN")

	// create logger
	var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)

	// connect to db
	db, err := openDB(cfg)
	if err!=nil {
		log.Fatal(err)
	}
	defer db.Close()

	logger.Println("database connected")

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
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	err = srv.ListenAndServe()
	logger.Fatalln(err)
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
