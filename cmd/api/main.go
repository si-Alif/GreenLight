package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	_ "github.com/lib/pq"
)

const version = "1.0.0"

// To store apps config data
type config struct {
	port int // port to listen on
	env string // development or production or test
	db struct{
		dsn string
	}

}

// this application struct will contain all dependencies / packages used in our application in a central place
type application struct{
	config config
	logger *slog.Logger

}

func main(){
	var cfg config

	flag.IntVar(&cfg.port , "port" , 4000 , "API server port")

	/*
	  flag parsing stages :
			1. Declare a flag with name , default values and a message , which initially returns a pointer pointing to the default value
			2. Now when the code encounters flag.Parse() it will start evaluating and parsing the command line args
			3. After this step the flag variable will be updated to point towards the value passed in the command line
	*/

	// cfg.env = *flag.String("env" , "development" , "Environment (development | production | test)")
	// ❌❌❌ This is wrong cause your setting "cfg.env"'s value to the default one as the code hasn't went to flag.Parse() yet

	env := flag.String("env" , "development" , "Environment (development | production | test)")

	flag.StringVar(&cfg.db.dsn , "db-dsn" , "postgres://greenlight:pa55word@localhost/greenlight" , "PostgreSQL DSN")

	flag.Parse()

	cfg.env = *env // ✅ This is now correct as env flag's now pointing to the passed variable value from the CLI

	logger := slog.New(slog.NewTextHandler(os.Stdout , nil))

	// open a connection pool
	db , err := openDB(cfg)

	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	defer db.Close()

	logger.Info("database connection pool established")

	app := &application{
		config: cfg,
		logger : logger,
	}

	srv := &http.Server{
		Addr: fmt.Sprintf(":%d" , cfg.port), // made a mistake here , remember Addr takes port pattern like :4000 and you forgot the colon
		Handler: app.routes(),
		IdleTimeout: time.Minute,
		ReadTimeout: time.Second * 5,
		WriteTimeout: time.Second * 10,
		ErrorLog: slog.NewLogLogger(logger.Handler() , slog.LevelError),
	}

	logger.Info("starting server " , "addr" , srv.Addr , "env" , cfg.env)

	err= srv.ListenAndServe()
	logger.Error(err.Error())
	os.Exit(1)

}

func openDB(cfg config) (*sql.DB , error){
	db ,err := sql.Open("postgres" , cfg.db.dsn) // opens a connection pool , not the connection itself , cz connection is done lazily(only when a request comes in) . So , we need to make sure all is ok so that we don't fell in some runtime error while connecting to the db which is done by using db.PingContext()

	if err != nil {
		return nil , err
	}

	ctx , cancel := context.WithTimeout(context.Background() , 5 * time.Second)
	defer cancel()

	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil , err
	}

	return  db , nil
}