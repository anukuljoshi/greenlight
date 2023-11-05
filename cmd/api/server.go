package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	// create a server with config
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}
	shutdownError := make(chan error)
	go func () {
		// create a quit channel which carries os.Signal values
		quit := make(chan os.Signal, 1)
		// user signal.Notify to listen fo SIGTERM, SIGINT signal and relay then to quit channel
		signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
		// read a signal from channel
		// this code will block until a signal is received
		s := <-quit
		// log a message when signal is caught
		app.logger.PrintInfo("shutting down server", map[string]string{
			"signal": s.String(),
		})
		// create a context with 5 second timeout
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		// call shutdown passing the ctx and catching error
		err := srv.Shutdown(ctx)
		if err!=nil {
			shutdownError <- err
		}
		// log message indicating background task are being completed
		app.logger.PrintInfo("completing background tasks", map[string]string{
			"addr": srv.Addr,
		})
		// call wait on app.wg to wait for all goroutine to complete
		app.wg.Wait()
		shutdownError <- nil
	}()
	app.logger.PrintInfo("starting server", map[string]string{
		"addr": srv.Addr,
		"env": app.config.env,
	})
	// calling shutdown return ErrServerClosed error
	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	// wait for shutdownError on channel
	err = <-shutdownError
	if err!=nil {
		return err
	}
	// graceful shutdown was successful
	app.logger.PrintInfo("stopped server", map[string]string{
		"addr": srv.Addr,
	})
	return nil
}
