package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/facebookgo/inject"
	"github.com/spf13/cobra"
	"github.com/hcnet/go/clients/federation"
	hc "github.com/hcnet/go/clients/auroraclient"
	"github.com/hcnet/go/clients/hcnettoml"
	"github.com/hcnet/go/services/bridge/internal/config"
	"github.com/hcnet/go/services/bridge/internal/db"
	"github.com/hcnet/go/services/bridge/internal/handlers"
	"github.com/hcnet/go/services/bridge/internal/listener"
	"github.com/hcnet/go/services/bridge/internal/submitter"
	supportConfig "github.com/hcnet/go/support/config"
	"github.com/hcnet/go/support/db/schema"
	"github.com/hcnet/go/support/errors"
	supportHttp "github.com/hcnet/go/support/http"
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
		Use:   "bridge",
		Short: "hcnet bridge server",
		Long:  `hcnet bridge server`,
		Run:   run,
	}

	rootCmd.Flags().BoolVarP(&migrateFlag, "migrate-db", "", false, "migrate DB to the newest schema version")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "bridge.cfg", "path to config file")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "displays bridge server version")
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

	if config.Database != nil {
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
		fmt.Printf("Bridge Server Version: %s \n", version)
		os.Exit(0)
		return
	}

	if len(config.APIKey) > 0 && len(config.APIKey) < 15 {
		err = errors.New("api-key have to be at least 15 chars long")
		return
	}

	requestHandler := handlers.RequestHandler{}

	httpClientWithTimeout := http.Client{
		Timeout: 60 * time.Second,
	}

	h := hc.Client{
		AuroraURL: config.Aurora,
		HTTP:       http.DefaultClient,
		AppName:    "bridge-server",
	}

	log.Print("Creating and initializing TransactionSubmitter")
	ts := submitter.NewTransactionSubmitter(&h, &database, config.NetworkPassphrase, time.Now)
	if err != nil {
		return
	}

	if config.Accounts.AuthorizingSeed == "" {
		log.Warning("No accounts.authorizing_seed param. Skipping...")
	} else {
		log.Print("Initializing Authorizing account")
		err = ts.InitAccount(config.Accounts.AuthorizingSeed)
		if err != nil {
			return
		}
	}

	if config.Accounts.BaseSeed == "" {
		log.Warning("No accounts.base_seed param. Skipping...")
	} else {
		log.Print("Initializing Base account")
		err = ts.InitAccount(config.Accounts.BaseSeed)
		if err != nil {
			return
		}
	}

	log.Print("TransactionSubmitter created")

	log.Print("Creating and starting PaymentListener")

	var paymentListener listener.PaymentListener

	if config.Accounts.ReceivingAccountID == "" {
		log.Warning("No accounts.receiving_account_id param. Skipping...")
	} else if config.Callbacks.Receive == "" {
		log.Warning("No callbacks.receive param. Skipping...")
	} else {
		paymentListener, err = listener.NewPaymentListener(&config, &database, &h, time.Now)
		if err != nil {
			return
		}
		err = paymentListener.Listen()
		if err != nil {
			return
		}

		log.Print("PaymentListener created")
	}

	hcnettomlClient := hcnettoml.Client{
		HTTP: &httpClientWithTimeout,
	}

	federationClient := federation.Client{
		HTTP:        &httpClientWithTimeout,
		HcNetTOML: &hcnettomlClient,
	}

	err = g.Provide(
		&inject.Object{Value: &requestHandler},
		&inject.Object{Value: &config},
		&inject.Object{Value: &hcnettomlClient},
		&inject.Object{Value: &federationClient},
		&inject.Object{Value: &h},
		&inject.Object{Value: &database},
		&inject.Object{Value: &ts},
		&inject.Object{Value: &paymentListener},
		&inject.Object{Value: &httpClientWithTimeout},
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
	mux := supportHttp.NewAPIMux(false)

	// Middlewares
	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	mux.Use(supportHttp.StripTrailingSlashMiddleware("/admin"))
	mux.Use(supportHttp.HeadersMiddleware(headers, "/admin/"))

	if a.config.APIKey != "" {
		mux.Use(apiKeyMiddleware(a.config.APIKey))
	}

	if a.config.Accounts.AuthorizingSeed != "" {
		mux.Post("/authorize", a.requestHandler.Authorize)
	} else {
		log.Warning("accounts.authorizing_seed not provided. /authorize endpoint will not be available.")
	}

	mux.Post("/create-keypair", a.requestHandler.CreateKeypair)
	mux.Post("/builder", a.requestHandler.Builder)
	mux.Post("/payment", a.requestHandler.Payment)
	mux.Get("/payment", a.requestHandler.Payment)
	mux.Post("/reprocess", a.requestHandler.Reprocess)

	mux.Get("/admin/received-payments", a.requestHandler.AdminReceivedPayments)
	mux.Get("/admin/received-payments/{id}", a.requestHandler.AdminReceivedPayment)
	mux.Get("/admin/sent-transactions", a.requestHandler.AdminSentTransactions)

	supportHttp.Run(supportHttp.Config{
		ListenAddr: fmt.Sprintf(":%d", *a.config.Port),
		Handler:    mux,
		OnStarting: func() {
			log.Infof("starting bridge server")
			log.Infof("listening on %d", *a.config.Port)
		},
	})
}

// apiKeyMiddleware checks for apiKey in a request and writes http.StatusForbidden if it's incorrect.
func apiKeyMiddleware(apiKey string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			k := r.PostFormValue("apiKey")
			if k != apiKey {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}
