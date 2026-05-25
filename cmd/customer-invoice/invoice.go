package main

import (
	"flag"
	"log"
	"os"
	"strconv"
	"strings"
)

const version = "1.0.0"

type config struct {
	port int
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

func main() {

	var cfg config

	flag.IntVar(&cfg.port, "port", 5000, "Server port to listen on")
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
}
