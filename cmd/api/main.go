package main

import (
	"fmt"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const version = "1.0.0"

// To store apps config data
type config struct {
	port int // port to listen on
	env string // development or production or test
}

// this application struct will contain all dependencies / packages used in our application in a central place
type application struct{
	config config
	logger *slog.Logger
}

func main(){
	var cfg config

	flag.IntVar(&cfg.port , "port" , 4000 , "API server port")
	// cfg.env = *flag.String("env" , "development" , "Environment (development | production | test)")
	flag.StringVar(&cfg.env , "env" , "development" , "Environment (development | production | test)")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout , nil))

	app := &application{
		config: cfg,
		logger : logger,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/v1/healthcheck" , app.healthcheckHandler)

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d" , cfg.port),
		Handler: mux,
		IdleTimeout: time.Minute,
		ReadTimeout: time.Second * 5,
		WriteTimeout: time.Second * 10,
		ErrorLog: slog.NewLogLogger(logger.Handler() , slog.LevelError),
	}

	logger.Info("starting server " , "addr" , srv.Addr , "env" , cfg.env)

	err:= srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)

}