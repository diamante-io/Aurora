package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/facebookgo/inject"
	"github.com/goji/httpauth"
	"github.com/spf13/cobra"
	"github.com/diamnet/go/clients/federation"
	"github.com/diamnet/go/clients/diamnettoml"
	"github.com/diamnet/go/services/compliance/internal/config"
	"github.com/diamnet/go/services/compliance/internal/crypto"
	"github.com/diamnet/go/services/compliance/internal/db"
	"github.com/diamnet/go/services/compliance/internal/handlers"
	supportConfig "github.com/diamnet/go/support/config"
	"github.com/diamnet/go/support/db/schema"
	"github.com/diamnet/go/support/errors"
	supportHttp "github.com/diamnet/go/support/http"
)

var app *App
var rootCmd *cobra.Command
var migrateFlag bool
var configFile string
var versionFlag bool
var version = "N/A"

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	rootCmd.Execute()
}

func init() {
	rootCmd = &cobra.Command{
		Use:   "compliance",
		Short: "diamnet compliance server",
		Long:  `diamnet compliance server`,
		Run:   run,
	}

	rootCmd.Flags().BoolVarP(&migrateFlag, "migrate-db", "", false, "migrate DB to the newest schema version")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "compliance.cfg", "path to config file")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "displays compliance server version")
}

func run(cmd *cobra.Command, args []string) {
	var cfg config.Config

	err := supportConfig.Read(configFile, &cfg)
	if err != nil {
		switch cause := errors.Cause(err).(type) {
		case *supportConfig.InvalidConfigError:
			log.Error("config file: ", cause)
		default:
			log.Error(err)
		}
		os.Exit(-1)
	}

	err = cfg.Validate()
	if err != nil {
		log.Fatal(err.Error())
		return
	}

	if cfg.LogFormat == "json" {
		log.SetFormatter(&log.JSONFormatter{})
	}

	app, err = NewApp(cfg, migrateFlag, versionFlag, version)

	if err != nil {
		log.Fatal(err.Error())
		return
	}

	app.Serve()
}

// App is the application object
type App struct {
	config         config.Config
	requestHandler handlers.RequestHandler
}

// NewApp constructs an new App instance from the provided config.
func NewApp(config config.Config, migrateFlag bool, versionFlag bool, version string) (app *App, err error) {
	var g inject.Graph

	var database db.PostgresDatabase

	if config.Database.URL != "" {
		err = database.Open(config.Database.URL)
		if err != nil {
			err = fmt.Errorf("Cannot connect to a DB: %s", err)
			return
		}
	}

	if migrateFlag {
		var migrationsApplied int
		migrationsApplied, err = schema.Migrate(
			database.GetDB(),
			db.Migrations,
			schema.MigrateUp,
			0,
		)
		if err != nil {
			return
		}

		log.Info("Applied migrations: ", migrationsApplied)
		os.Exit(0)
		return
	}

	if versionFlag {
		fmt.Printf("Compliance Server Version: %s \n", version)
		os.Exit(0)
		return
	}

	requestHandler := handlers.RequestHandler{}

	httpClientWithTimeout := http.Client{
		Timeout: 10 * time.Second,
	}

	diamnettomlClient := diamnettoml.Client{
		HTTP: &httpClientWithTimeout,
	}

	federationClient := federation.Client{
		HTTP:        &httpClientWithTimeout,
		DiamNetTOML: &diamnettomlClient,
	}

	err = g.Provide(
		&inject.Object{Value: &requestHandler},
		&inject.Object{Value: &config},
		&inject.Object{Value: &database},
		&inject.Object{Value: &crypto.SignerVerifier{}},
		&inject.Object{Value: &diamnettomlClient},
		&inject.Object{Value: &federationClient},
		&inject.Object{Value: &httpClientWithTimeout},
		&inject.Object{Value: &handlers.NonceGenerator{}},
	)

	if err != nil {
		log.Fatal("Injector: ", err)
	}

	if err := g.Populate(); err != nil {
		log.Fatal("Injector: ", err)
	}

	app = &App{
		config:         config,
		requestHandler: requestHandler,
	}
	return
}

// Serve starts the server
func (a *App) Serve() {
	// External endpoints
	external := supportHttp.NewAPIMux(false)

	// Middlewares
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")

	external.Use(supportHttp.StripTrailingSlashMiddleware())
	external.Use(supportHttp.HeadersMiddleware(headers))

	external.Post("/", a.requestHandler.HandlerAuth)
	if a.config.TxStatusAuth != nil {
		external.Method("GET", "/tx_status", httpauth.SimpleBasicAuth(a.config.TxStatusAuth.Username, a.config.TxStatusAuth.Password)(http.HandlerFunc(a.requestHandler.HandlerTxStatus)))
	}
	go func() {
		supportHttp.Run(supportHttp.Config{
			ListenAddr: fmt.Sprintf(":%d", *a.config.ExternalPort),
			Handler:    external,
			TLS:        a.config.TLS,
			OnStarting: func() {
				log.Infof("External server listening on %d", *a.config.ExternalPort)
			},
		})
	}()

	// Internal endpoints
	internal := supportHttp.NewAPIMux(false)

	internal.Use(supportHttp.StripTrailingSlashMiddleware("/admin"))
	internal.Use(supportHttp.HeadersMiddleware(headers, "/admin/"))

	internal.Post("/send", a.requestHandler.HandlerSend)
	internal.Post("/receive", a.requestHandler.HandlerReceive)
	internal.Post("/allow_access", a.requestHandler.HandlerAllowAccess)
	internal.Post("/remove_access", a.requestHandler.HandlerRemoveAccess)

	supportHttp.Run(supportHttp.Config{
		ListenAddr: fmt.Sprintf(":%d", *a.config.InternalPort),
		Handler:    internal,
		OnStarting: func() {
			log.Infof("Internal server listening on %d", *a.config.InternalPort)
		},
	})
}
