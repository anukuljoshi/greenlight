package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// config struct to hold settings for our application
type config struct {
	port int
	env string
}

// application struct to hold dependencies for handlers, middlewares, helpers
type application struct {
	config config
	logger *log.Logger
}

func main() {
	var cfg config

	flag.IntVar(&cfg.port, "port", 4000, "PORT for application")
	flag.StringVar(&cfg.env, "env", "development", "Environment (development|staging|production)")
	flag.Parse()

	var logger = log.New(os.Stdout, "", log.Ldate|log.Ltime)
	var app = &application{
		config: cfg,
		logger: logger,
	}

	var srv = &http.Server{
		Addr: fmt.Sprintf(":%d", cfg.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Printf("starting %s server on %s", cfg.env, srv.Addr)
	var err = srv.ListenAndServe()
	logger.Fatalln(err)
}
