package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port), // made a mistake here , remember Addr takes port pattern like :4000 and you forgot the colon
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  time.Second * 5,
		WriteTimeout: time.Second * 10,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	// make a channel to receive any errors raised by the Shutdown() method
	shutdownErr := make(chan error)

	// persistent listening to end of process signal in a background goroutine
	go func() {
		quit := make(chan os.Signal, 1)

		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		// quit if notification detected
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		// define a context with 30s timeout
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		defer cancel()

		// shut down the server with the context window of 30s we just created
		// if shutdown is successful , then Shutdown() returns nil
		// or in case any error occurs or shutdown wasn't complete in the 30s time frame then it will return error which will be relayed to shutdownErr channel
		shutdownErr <- srv.Shutdown(ctx)

	}()

	app.logger.Info("starting server ", "addr", srv.Addr, "env", app.config.env)

	// Start the server. ListenAndServe() will return http.ErrServerClosed
	// when Shutdown() is called, which is expected behavior during graceful shutdown
	err := srv.ListenAndServe()

	// If error is NOT http.ErrServerClosed, it means something unexpected happened
	// (e.g., port already in use, permission denied, etc.)
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	// If we reach here, it means Shutdown() was called (we got http.ErrServerClosed)
	// Now we wait for the graceful shutdown to complete by reading from shutdownErr channel
	// This blocks until the goroutine sends the result of srv.Shutdown(ctx)
	err = <-shutdownErr
	if err != nil {
		return err
	}

	// Graceful shutdown completed successfully
	app.logger.Info("stopped server", "addr", srv.Addr)

	return nil
}
