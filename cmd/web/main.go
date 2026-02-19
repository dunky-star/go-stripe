package main

import (
    "flag"
    "fmt"
    "html/template"
    "log"
    "net/http"
    "os"
    "time"
)

const (
    version    = "1.0.0"
    cssVersion = "1"
)

type config struct {
	port int
	env string
	api string
	db struct {
		dsn string
	}
	stripe struct {
		secret string
		key string
	}
}

type application struct {
	config config
	infoLog *log.Logger
	errorLog *log.Logger
	templateCache map[string]*template.Template
	version string
}

func (app *application) serve() error {
	// Create a new http.Server struct. We can specify any non-default values for the fields in this struct.
	srv := &http.Server{
		Addr: fmt.Sprintf(":%d", app.config.port),
		Handler: app.routes(),
		IdleTimeout: 30 *time.Minute,
		ReadTimeout: 10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	app.infoLog.Printf("Starting API server in %s mode on port %d", app.config.env, app.config.port)
	err := srv.ListenAndServe()
    return err
}

func main() {
	// Initialize a new instance of the config struct.
	var cfg config
	
	// Read the value of the PORT environment variable into the config struct. If it’s not set, default to 4000.
	flag.IntVar(&cfg.port, "port", 4000, "API server port to listen on")
	flag.StringVar(&cfg.env, "env", "development", "Environment (dev|staging|prod)")
	flag.StringVar(&cfg.api, "api", "http://localhost:4001", "API server URL")
	flag.Parse()

	// Foe security, we should not store our Stripe secret key in the config struct. 
	// Instead, we can read it from an environment variable.
	cfg.stripe.secret = os.Getenv("STRIPE_SECRET")
	cfg.stripe.key = os.Getenv("STRIPE_KEY")

	// Logging setup
	infoLog := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errorLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime|log.Lshortfile)

	// Template caching setup
	tc := make(map[string]*template.Template)

	// Initialize a new instance of the application struct, 
	// containing the config struct and the loggers.
	app := &application{
		config: cfg,
		infoLog: infoLog,
		errorLog: errorLog,
		templateCache: tc,
		version: version,
	}
	// Call the serve() method on our application struct.
	err := app.serve()
	if err != nil {
		app.errorLog.Printf("Error starting server: %v", err)
		log.Fatal(err)
	}
}