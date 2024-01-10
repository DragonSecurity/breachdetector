package main

import (
	"flag"
	"log/slog"
	"net/http"
	"os"
	"runtime/debug"
	"sync"

	"github.com/dragonsecurity/breachdetector/internal/database"
	"github.com/dragonsecurity/breachdetector/internal/smtp"

	"github.com/carlmjohnson/versioninfo"
	"github.com/gorilla/sessions"
	"github.com/lmittmann/tint"
)

func main() {
	logger := slog.New(tint.NewHandler(os.Stdout, &tint.Options{Level: slog.LevelDebug}))

	err := run(logger)
	if err != nil {
		trace := string(debug.Stack())
		logger.Error(err.Error(), "trace", trace)
		os.Exit(1)
	}
}

type config struct {
	baseURL   string
	httpPort  int
	basicAuth struct {
		username       string
		hashedPassword string
	}
	autoHTTPS struct {
		domain  string
		email   string
		staging bool
	}
	cookie struct {
		secretKey string
	}
	db struct {
		dsn         string
		automigrate bool
	}
	notifications struct {
		email string
	}
	session struct {
		secretKey    string
		oldSecretKey string
	}
	smtp struct {
		host     string
		port     int
		username string
		password string
		from     string
	}
}

type application struct {
	config       config
	db           *database.DB
	logger       *slog.Logger
	mailer       *smtp.Mailer
	sessionStore *sessions.CookieStore
	wg           sync.WaitGroup
}

func run(logger *slog.Logger) error {
	var cfg config

	flag.StringVar(&cfg.baseURL, "base-url", "http://localhost:4444", "base URL for the application")
	flag.IntVar(&cfg.httpPort, "http-port", 4444, "port to listen on for HTTP requests")
	flag.StringVar(&cfg.autoHTTPS.domain, "auto-https-domain", "", "domain to enable automatic HTTPS for")
	flag.StringVar(&cfg.autoHTTPS.email, "auto-https-email", "admin@example.com", "contact email address for problems with LetsEncrypt certificates")
	flag.BoolVar(&cfg.autoHTTPS.staging, "auto-https-staging", false, "use LetsEncrypt staging environment")
	flag.StringVar(&cfg.basicAuth.username, "basic-auth-username", "admin", "basic auth username")
	flag.StringVar(&cfg.basicAuth.hashedPassword, "basic-auth-hashed-password", "$2a$10$jRb2qniNcoCyQM23T59RfeEQUbgdAXfR6S0scynmKfJa5Gj3arGJa", "basic auth password hashed with bcrpyt")
	flag.StringVar(&cfg.cookie.secretKey, "cookie-secret-key", "p44v4phzhji3mk6z5qzlc4h6tves666n", "secret key for cookie authentication/encryption")
	flag.StringVar(&cfg.db.dsn, "db-dsn", "db.sqlite", "sqlite3 DSN")
	flag.BoolVar(&cfg.db.automigrate, "db-automigrate", true, "run migrations on startup")
	flag.StringVar(&cfg.notifications.email, "notifications-email", "", "contact email address for error notifications")
	flag.StringVar(&cfg.session.secretKey, "session-secret-key", "4nhkdzwbbhpbmwqq267azik6jg3btamo", "secret key for session cookie authentication")
	flag.StringVar(&cfg.session.oldSecretKey, "session-old-secret-key", "", "previous secret key for session cookie authentication")
	flag.StringVar(&cfg.smtp.host, "smtp-host", "example.smtp.host", "smtp host")
	flag.IntVar(&cfg.smtp.port, "smtp-port", 25, "smtp port")
	flag.StringVar(&cfg.smtp.username, "smtp-username", "example_username", "smtp username")
	flag.StringVar(&cfg.smtp.password, "smtp-password", "pa55word", "smtp password")
	flag.StringVar(&cfg.smtp.from, "smtp-from", "Example Name <no-reply@example.org>", "smtp sender")

	versioninfo.AddFlag(nil)

	flag.Parse()

	db, err := database.New(cfg.db.dsn, cfg.db.automigrate)
	if err != nil {
		return err
	}
	defer db.Close()

	mailer, err := smtp.NewMailer(cfg.smtp.host, cfg.smtp.port, cfg.smtp.username, cfg.smtp.password, cfg.smtp.from)
	if err != nil {
		return err
	}

	keyPairs := [][]byte{[]byte(cfg.session.secretKey), nil}
	if cfg.session.oldSecretKey != "" {
		keyPairs = append(keyPairs, []byte(cfg.session.oldSecretKey), nil)
	}

	sessionStore := sessions.NewCookieStore(keyPairs...)
	sessionStore.Options = &sessions.Options{
		HttpOnly: true,
		MaxAge:   86400 * 7,
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   true,
	}

	app := &application{
		config:       cfg,
		db:           db,
		logger:       logger,
		mailer:       mailer,
		sessionStore: sessionStore,
	}

	if cfg.autoHTTPS.domain != "" {
		return app.serveAutoHTTPS()
	}

	return app.serveHTTP()
}
