package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

const version = "1.0.0"

type config struct {
	port int
	env  string
	smtp struct {
		host     string
		port     int
		username string
		password string
	}
	frontend string
}

type application struct {
	config   config
	infoLog  *log.Logger
	errorLog *log.Logger
	version  string
}

func (app *application) serve() error {
	// Create a new http.Server struct. We can specify any non-default values for the fields in this struct.
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", app.config.port),
		Handler:           app.routes(),
		IdleTimeout:       30 * time.Minute,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      5 * time.Second,
	}
	app.infoLog.Printf("Starting API server in %s mode on port %d", app.config.env, app.config.port)
	err := srv.ListenAndServe()
	return err
}

func main() {

	_ = godotenv.Load()

	var cfg config

	flag.IntVar(&cfg.port, "port", 5001, "Server port to listen on")
	flag.Parse()

	cfg.smtp.host = strings.TrimSpace(os.Getenv("SMTP_HOST"))
	if p := strings.TrimSpace(os.Getenv("SMTP_PORT")); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cfg.smtp.port = n
		}
	}
	cfg.smtp.username = os.Getenv("SMTP_USER")
	cfg.smtp.password = os.Getenv("SMTP_PASSWORD")

	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stdout, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	app := &application{
		config:   cfg,
		infoLog:  infoLog,
		errorLog: errorLog,
		version:  version,
	}

	app.CreateDirIfNotExist("./invoices")

	// Call the serve() method on our application struct.
	err := app.serve()
	if err != nil {
		app.errorLog.Printf("Error starting server: %v", err)
		log.Fatal(err)
	}
}
