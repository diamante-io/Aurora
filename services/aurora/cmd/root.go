package cmd

import (
	"fmt"
	"go/types"
	stdLog "log"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	aurora "github.com/hcnet/go/services/aurora/internal"
	"github.com/hcnet/go/services/aurora/internal/db2/schema"
	apkg "github.com/hcnet/go/support/app"
	support "github.com/hcnet/go/support/config"
	"github.com/hcnet/go/support/log"
	"github.com/throttled/throttled"
)

var config aurora.Config

var rootCmd = &cobra.Command{
	Use:   "aurora",
	Short: "client-facing api server for the hcnet network",
	Long:  "client-facing api server for the hcnet network. It acts as the interface between HcNet Core and applications that want to access the HcNet network. It allows you to submit transactions to the network, check the status of accounts, subscribe to event streams and more.",
	Run: func(cmd *cobra.Command, args []string) {
		initApp().Serve()
	},
}

// validateBothOrNeither ensures that both options are provided, if either is provided.
func validateBothOrNeither(option1, option2 string) {
	arg1, arg2 := viper.GetString(option1), viper.GetString(option2)
	if arg1 != "" && arg2 == "" {
		stdLog.Fatalf("Invalid config: %s = %s, but corresponding option %s is not configured", option1, arg1, option2)
	}
	if arg1 == "" && arg2 != "" {
		stdLog.Fatalf("Invalid config: %s = %s, but corresponding option %s is not configured", option2, arg2, option1)
	}
}

// checkMigrations looks for necessary database migrations and fails with a descriptive error if migrations are needed.
func checkMigrations() {
	migrationsToApplyUp := schema.GetMigrationsUp(viper.GetString("db-url"))
	if len(migrationsToApplyUp) > 0 {
		stdLog.Printf(`There are %v migrations to apply in the "up" direction.`, len(migrationsToApplyUp))
		stdLog.Printf("The necessary migrations are: %v", migrationsToApplyUp)
		stdLog.Printf("A database migration is required to run this version (%v) of Aurora. Run \"aurora db migrate up\" to update your DB. Consult the Changelog (https://github.com/hcnet/go/blob/master/services/aurora/CHANGELOG.md) for more information.", apkg.Version())
		os.Exit(1)
	}

	nMigrationsDown := schema.GetNumMigrationsDown(viper.GetString("db-url"))
	if nMigrationsDown > 0 {
		stdLog.Printf("A database migration DOWN to an earlier version of the schema is required to run this version (%v) of Aurora. Consult the Changelog (https://github.com/hcnet/go/blob/master/services/aurora/CHANGELOG.md) for more information.", apkg.Version())
		stdLog.Printf("In order to migrate the database DOWN, using the HIGHEST version number of Aurora you have installed (not this binary), run \"aurora db migrate down %v\".", nMigrationsDown)
		os.Exit(1)
	}
}

// configOpts defines the complete flag configuration for aurora.
// Add a new entry here to connect a new field in the aurora.Config struct
var configOpts = []*support.ConfigOption{
	&support.ConfigOption{
		Name:      "db-url",
		EnvVar:    "DATABASE_URL",
		ConfigKey: &config.DatabaseURL,
		OptType:   types.String,
		Required:  true,
		Usage:     "aurora postgres database to connect with",
	},
	&support.ConfigOption{
		Name:      "hcnet-core-db-url",
		EnvVar:    "HCNET_CORE_DATABASE_URL",
		ConfigKey: &config.HcNetCoreDatabaseURL,
		OptType:   types.String,
		Required:  true,
		Usage:     "hcnet-core postgres database to connect with",
	},
	&support.ConfigOption{
		Name:      "hcnet-core-url",
		ConfigKey: &config.HcNetCoreURL,
		OptType:   types.String,
		Required:  true,
		Usage:     "hcnet-core to connect with (for http commands)",
	},
	&support.ConfigOption{
		Name:        "history-archive-urls",
		ConfigKey:   &config.HistoryArchiveURLs,
		OptType:     types.String,
		Required:    false,
		FlagDefault: "",
		CustomSetValue: func(co *support.ConfigOption) {
			stringOfUrls := viper.GetString(co.Name)
			urlStrings := strings.Split(stringOfUrls, ",")

			*(co.ConfigKey.(*[]string)) = urlStrings
		},
		Usage: "comma-separated list of hcnet history archives to connect with",
	},
	&support.ConfigOption{
		Name:        "port",
		ConfigKey:   &config.Port,
		OptType:     types.Uint,
		FlagDefault: uint(8000),
		Usage:       "tcp port to listen on for http requests",
	},
	&support.ConfigOption{
		Name:        "max-db-connections",
		ConfigKey:   &config.MaxDBConnections,
		OptType:     types.Int,
		FlagDefault: 0,
		Usage:       "when set has a priority over aurora-db-max-open-connections, aurora-db-max-idle-connections, core-db-max-open-connections, core-db-max-idle-connections. max aurora database open connections. may need to be increased when responses are slow but DB CPU is normal",
	},
	&support.ConfigOption{
		Name:        "aurora-db-max-open-connections",
		ConfigKey:   &config.AuroraDBMaxOpenConnections,
		OptType:     types.Int,
		FlagDefault: 20,
		Usage:       "max aurora database open connections. may need to be increased when responses are slow but DB CPU is normal",
	},
	&support.ConfigOption{
		Name:        "aurora-db-max-idle-connections",
		ConfigKey:   &config.AuroraDBMaxIdleConnections,
		OptType:     types.Int,
		FlagDefault: 20,
		Usage:       "max aurora database idle connections. may need to be set to the same value as aurora-db-max-open-connections when responses are slow and DB CPU is normal, because it may indicate that a lot of time is spent closing/opening idle connections. This can happen in case of high variance in number of requests. must be equal or lower than max open connections",
	},
	&support.ConfigOption{
		Name:        "core-db-max-open-connections",
		ConfigKey:   &config.CoreDBMaxOpenConnections,
		OptType:     types.Int,
		FlagDefault: 20,
		Usage:       "max core database open connections. may need to be increased when responses are slow but DB CPU is normal",
	},
	&support.ConfigOption{
		Name:        "core-db-max-idle-connections",
		ConfigKey:   &config.CoreDBMaxIdleConnections,
		OptType:     types.Int,
		FlagDefault: 20,
		Usage:       "max core database idle connections. may need to be set to the same value as core-db-max-open-connections when responses are slow and DB CPU is normal, because it may indicate that a lot of time is spent closing/opening idle connections. This can happen in case of high variance in number of requests. must be equal or lower than max open connections",
	},
	&support.ConfigOption{
		Name:           "sse-update-frequency",
		ConfigKey:      &config.SSEUpdateFrequency,
		OptType:        types.Int,
		FlagDefault:    5,
		CustomSetValue: support.SetDuration,
		Usage:          "defines how often streams should check if there's a new ledger (in seconds), may need to increase in case of big number of streams",
	},
	&support.ConfigOption{
		Name:           "connection-timeout",
		ConfigKey:      &config.ConnectionTimeout,
		OptType:        types.Int,
		FlagDefault:    55,
		CustomSetValue: support.SetDuration,
		Usage:          "defines the timeout of connection after which 504 response will be sent or stream will be closed, if Aurora is behind a load balancer with idle connection timeout, this should be set to a few seconds less that idle timeout",
	},
	&support.ConfigOption{
		Name:        "per-hour-rate-limit",
		ConfigKey:   &config.RateQuota,
		OptType:     types.Int,
		FlagDefault: 3600,
		CustomSetValue: func(co *support.ConfigOption) {
			var rateLimit *throttled.RateQuota = nil
			perHourRateLimit := viper.GetInt(co.Name)
			if perHourRateLimit != 0 {
				rateLimit = &throttled.RateQuota{
					MaxRate:  throttled.PerHour(perHourRateLimit),
					MaxBurst: 100,
				}
				*(co.ConfigKey.(**throttled.RateQuota)) = rateLimit
			}
		},
		Usage: "max count of requests allowed in a one hour period, by remote ip address",
	},
	&support.ConfigOption{
		Name:      "rate-limit-redis-key",
		ConfigKey: &config.RateLimitRedisKey,
		OptType:   types.String,
		Usage:     "redis key for storing rate limit data, useful when deploying a cluster of Auroras, ignored when redis-url is empty",
	},
	&support.ConfigOption{
		Name:      "redis-url",
		ConfigKey: &config.RedisURL,
		OptType:   types.String,
		Usage:     "redis to connect with, for rate limiting",
	},
	&support.ConfigOption{
		Name:           "friendbot-url",
		ConfigKey:      &config.FriendbotURL,
		OptType:        types.String,
		CustomSetValue: support.SetURL,
		Usage:          "friendbot service to redirect to",
	},
	&support.ConfigOption{
		Name:        "log-level",
		ConfigKey:   &config.LogLevel,
		OptType:     types.String,
		FlagDefault: "info",
		CustomSetValue: func(co *support.ConfigOption) {
			ll, err := logrus.ParseLevel(viper.GetString(co.Name))
			if err != nil {
				stdLog.Fatalf("Could not parse log-level: %v", viper.GetString(co.Name))
			}
			*(co.ConfigKey.(*logrus.Level)) = ll
		},
		Usage: "minimum log severity (debug, info, warn, error) to log",
	},
	&support.ConfigOption{
		Name:      "log-file",
		ConfigKey: &config.LogFile,
		OptType:   types.String,
		Usage:     "name of the file where logs will be saved (leave empty to send logs to stdout)",
	},
	&support.ConfigOption{
		Name:        "max-path-length",
		ConfigKey:   &config.MaxPathLength,
		OptType:     types.Uint,
		FlagDefault: uint(4),
		Usage:       "the maximum number of assets on the path in `/paths` endpoint",
	},
	&support.ConfigOption{
		Name:      "network-passphrase",
		ConfigKey: &config.NetworkPassphrase,
		OptType:   types.String,
		Required:  true,
		Usage:     "Override the network passphrase",
	},
	&support.ConfigOption{
		Name:      "sentry-dsn",
		ConfigKey: &config.SentryDSN,
		OptType:   types.String,
		Usage:     "Sentry URL to which panics and errors should be reported",
	},
	&support.ConfigOption{
		Name:      "loggly-token",
		ConfigKey: &config.LogglyToken,
		OptType:   types.String,
		Usage:     "Loggly token, used to configure log forwarding to loggly",
	},
	&support.ConfigOption{
		Name:        "loggly-tag",
		ConfigKey:   &config.LogglyTag,
		OptType:     types.String,
		FlagDefault: "aurora",
		Usage:       "Tag to be added to every loggly log event",
	},
	&support.ConfigOption{
		Name:      "tls-cert",
		ConfigKey: &config.TLSCert,
		OptType:   types.String,
		Usage:     "TLS certificate file to use for securing connections to aurora",
	},
	&support.ConfigOption{
		Name:      "tls-key",
		ConfigKey: &config.TLSKey,
		OptType:   types.String,
		Usage:     "TLS private key file to use for securing connections to aurora",
	},
	&support.ConfigOption{
		Name:        "ingest",
		ConfigKey:   &config.Ingest,
		OptType:     types.Bool,
		FlagDefault: false,
		Usage:       "causes this aurora process to ingest data from hcnet-core into aurora's db",
	},
	&support.ConfigOption{
		Name:        "ingest-failed-transactions",
		ConfigKey:   &config.IngestFailedTransactions,
		OptType:     types.Bool,
		FlagDefault: false,
		Usage:       "causes this aurora process to ingest failed transactions data",
	},
	&support.ConfigOption{
		Name:        "cursor-name",
		EnvVar:      "CURSOR_NAME",
		ConfigKey:   &config.CursorName,
		OptType:     types.String,
		FlagDefault: "HORIZON",
		Usage:       "ingestor cursor used by aurora to ingest from hcnet core. must be uppercase and unique for each aurora instance ingesting from that core instance.",
	},
	&support.ConfigOption{
		Name:        "history-retention-count",
		ConfigKey:   &config.HistoryRetentionCount,
		OptType:     types.Uint,
		FlagDefault: uint(0),
		Usage:       "the minimum number of ledgers to maintain within aurora's history tables.  0 signifies an unlimited number of ledgers will be retained",
	},
	&support.ConfigOption{
		Name:        "history-stale-threshold",
		ConfigKey:   &config.StaleThreshold,
		OptType:     types.Uint,
		FlagDefault: uint(0),
		Usage:       "the maximum number of ledgers the history db is allowed to be out of date from the connected hcnet-core db before aurora considers history stale",
	},
	&support.ConfigOption{
		Name:        "skip-cursor-update",
		ConfigKey:   &config.SkipCursorUpdate,
		OptType:     types.Bool,
		FlagDefault: false,
		Usage:       "causes the ingester to skip reporting the last imported ledger state to hcnet-core",
	},
	&support.ConfigOption{
		Name:        "enable-asset-stats",
		ConfigKey:   &config.EnableAssetStats,
		OptType:     types.Bool,
		FlagDefault: false,
		Usage:       "enables asset stats during the ingestion and expose `/assets` endpoint, Enabling it has a negative impact on CPU",
	},
	&support.ConfigOption{
		Name:        "enable-experimental-ingestion",
		ConfigKey:   &config.EnableExperimentalIngestion,
		OptType:     types.Bool,
		FlagDefault: false,
		Usage:       "[EXPERIMENTAL] enables experimental ingestion system",
	},
	&support.ConfigOption{
		Name:        "ingest-state-reader-temp-set",
		ConfigKey:   &config.IngestStateReaderTempSet,
		OptType:     types.String,
		FlagDefault: "memory",
		Usage:       "defines where to store temporary objects during state ingestion: `memory` (default, more RAM usage, faster) or `postgres` (less RAM usage, slower)",
	},
}

func init() {
	for _, co := range configOpts {
		err := co.Init(rootCmd)
		if err != nil {
			stdLog.Fatal(err.Error())
		}
	}

	viper.BindPFlags(rootCmd.PersistentFlags())
}

func initApp() *aurora.App {
	initConfig()
	return aurora.NewApp(config)
}

func initConfig() {
	// Verify required options and load the config struct
	for _, co := range configOpts {
		co.Require()
		co.SetValue()
	}

	// Migrations should be checked as early as possible
	checkMigrations()

	// Validate options that should be provided together
	validateBothOrNeither("tls-cert", "tls-key")
	validateBothOrNeither("rate-limit-redis-key", "redis-url")

	// Configure log file
	if config.LogFile != "" {
		logFile, err := os.OpenFile(config.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err == nil {
			log.DefaultLogger.Logger.Out = logFile
		} else {
			stdLog.Fatalf("Failed to open file to log: %s", err)
		}
	}

	// Configure log level
	log.DefaultLogger.Logger.SetLevel(config.LogLevel)

	if config.IngestStateReaderTempSet != "memory" && config.IngestStateReaderTempSet != "postgres" {
		log.Fatal("Invalid `ingest-state-reader-temp-set` value: " + config.IngestStateReaderTempSet)
	}

	// Configure DB params. When config.MaxDBConnections is set, set other
	// DB params to that value for backward compatibility.
	if config.MaxDBConnections != 0 {
		config.AuroraDBMaxOpenConnections = config.MaxDBConnections
		config.AuroraDBMaxIdleConnections = config.MaxDBConnections
		config.CoreDBMaxOpenConnections = config.MaxDBConnections
		config.CoreDBMaxIdleConnections = config.MaxDBConnections
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
