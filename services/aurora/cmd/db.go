package cmd

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/diamnet/go/services/aurora/internal/db2/schema"
	"github.com/diamnet/go/services/aurora/internal/ingest"
	"github.com/diamnet/go/services/aurora/internal/util"
	"github.com/diamnet/go/support/db"
	"github.com/diamnet/go/support/errors"
	hlog "github.com/diamnet/go/support/log"
)

type reingestType int

const (
	byAll reingestType = iota
	byRange
	bySeq
	byOutdated
)

var dbCmd = &cobra.Command{
	Use:   "db [command]",
	Short: "commands to manage aurora's postgres db",
}

var dbBackfillCmd = &cobra.Command{
	Use:   "backfill [COUNT]",
	Short: "backfills aurora history for COUNT ledgers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			log.Println("Missing COUNT. Usage: backfill [COUNT].")
			return
		}

		initApp().UpdateLedgerState()

		i := ingestSystem(ingest.Config{
			IngestFailedTransactions: config.IngestFailedTransactions,
		})
		i.SkipCursorUpdate = true
		parsed, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			log.Fatal(err)
		}

		err = i.Backfill(uint(parsed))
		if err != nil {
			log.Fatal(err)
		}
	},
}

var dbInitAssetStatsCmd = &cobra.Command{
	Use:   "init-asset-stats",
	Short: "initializes values for assets stats",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		hdb, err := db.Open("postgres", config.DatabaseURL)
		if err != nil {
			log.Fatal(err)
		}

		cdb, err := db.Open("postgres", config.DiamNetCoreDatabaseURL)
		if err != nil {
			log.Fatal(err)
		}

		assetStats := ingest.AssetStats{
			CoreSession:    cdb,
			HistorySession: hdb,
		}

		log.Println("Getting assets from core DB...")

		count, err := assetStats.AddAllAssetsFromCore()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Updating %d assets...\n", count)

		err = assetStats.UpdateAssetStats()
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Added stats for %d assets...\n", count)
	},
}

var dbClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "clears all imported historical data",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		err := ingestSystem(ingest.Config{
			IngestFailedTransactions: config.IngestFailedTransactions,
		}).ClearAll()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var dbInitCmd = &cobra.Command{
	Use:   "init",
	Short: "install schema",
	Long:  "init initializes the postgres database used by aurora.",
	Run: func(cmd *cobra.Command, args []string) {
		dbConn, err := db.Open("postgres", viper.GetString("db-url"))
		if err != nil {
			log.Fatal(err)
		}

		err = schema.Init(dbConn)
		if err != nil {
			log.Fatal(err)
		}
	},
}

var dbMigrateCmd = &cobra.Command{
	Use:   "migrate [up|down|redo] [COUNT]",
	Short: "migrate schema",
	Long:  "performs a schema migration command",
	Run: func(cmd *cobra.Command, args []string) {
		// Allow invokations with 1 or 2 args.  All other args counts are erroneous.
		if len(args) < 1 || len(args) > 2 {
			cmd.Usage()
			os.Exit(1)
		}

		dir := schema.MigrateDir(args[0])
		count := 0

		// If a second arg is present, parse it to an int and use it as the count
		// argument to the migration call.
		if len(args) == 2 {
			var err error
			count, err = strconv.Atoi(args[1])
			if err != nil {
				log.Println(err)
				cmd.Usage()
				os.Exit(1)
			}
		}

		db, err := sql.Open("postgres", viper.GetString("db-url"))
		if err != nil {
			log.Fatal(err)
		}

		numMigrationsRun, err := schema.Migrate(db, dir, count)
		if err != nil {
			log.Fatal(err)
		}

		if numMigrationsRun == 0 {
			log.Println("No migrations applied.")
		} else {
			log.Printf("Successfully applied %d migrations.\n", numMigrationsRun)
		}
	},
}

var dbReapCmd = &cobra.Command{
	Use:   "reap",
	Short: "reaps (i.e. removes) any reapable history data",
	Long:  "reap removes any historical data that is earlier than the configured retention cutoff",
	Run: func(cmd *cobra.Command, args []string) {
		err := initApp().DeleteUnretainedHistory()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var dbRebaseCmd = &cobra.Command{
	Use:   "rebase",
	Short: "rebases clears the aurora db and ingests the latest ledger segment from diamnet-core",
	Long:  "...",
	Run: func(cmd *cobra.Command, args []string) {
		initConfig()

		i := ingestSystem(ingest.Config{
			IngestFailedTransactions: config.IngestFailedTransactions,
		})
		i.SkipCursorUpdate = true

		err := i.RebaseHistory()
		if err != nil {
			log.Fatal(err)
		}
	},
}

var dbReingestCmd = &cobra.Command{
	Use:   "reingest [Ledger sequence numbers (leave it empty for reingesting from the very beginning)]",
	Short: "reingest all ledgers or ledgers specified by individual sequence numbers",
	Long:  "reingest runs the ingestion pipeline over every ledger or ledgers specified by individual sequence numbers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			reingest(byAll)
		} else {
			argsInt32 := make([]int32, 0, len(args))
			for _, arg := range args {
				seq, err := strconv.Atoi(arg)
				if err != nil {
					cmd.Usage()
					log.Fatalf(`Invalid sequence number "%s"`, arg)
				}
				argsInt32 = append(argsInt32, int32(seq))
			}

			reingest(bySeq, argsInt32...)
		}
	},
}

var dbReingestRangeCmd = &cobra.Command{
	Use:   "range [Start sequence number] [End sequence number]",
	Short: "reingests ledgers within a range",
	Long:  "reingests ledgers between X and Y sequence number (closed intervals)",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 2 {
			cmd.Usage()
			os.Exit(1)
		}

		argsInt32 := make([]int32, 0, len(args))
		for _, arg := range args {
			seq, err := strconv.Atoi(arg)
			if err != nil {
				cmd.Usage()
				log.Fatalf(`Invalid sequence number "%s"`, arg)
			}
			argsInt32 = append(argsInt32, int32(seq))
		}

		reingest(byRange, argsInt32...)
	},
}

var dbReingestOutdatedCmd = &cobra.Command{
	Use:   "outdated",
	Short: "reingests all outdated ledgers",
	Long:  "reingests ledgers whose version is less than the current version up to a million ledgers",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) > 0 {
			log.Println("ignoring args...")
		}

		reingest(byOutdated)
	},
}

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(
		dbInitCmd,
		dbInitAssetStatsCmd,
		dbBackfillCmd,
		dbClearCmd,
		dbMigrateCmd,
		dbReapCmd,
		dbReingestCmd,
		dbRebaseCmd,
	)
	dbReingestCmd.AddCommand(dbReingestRangeCmd, dbReingestOutdatedCmd)
}

func ingestSystem(ingestConfig ingest.Config) *ingest.System {
	hdb, err := db.Open("postgres", config.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	cdb, err := db.Open("postgres", config.DiamNetCoreDatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	passphrase := viper.GetString("network-passphrase")
	if passphrase == "" {
		log.Fatal("network-passphrase is blank: reingestion requires manually setting passphrase")
	}

	return ingest.New(passphrase, config.DiamNetCoreURL, cdb, hdb, ingestConfig)
}

func reingest(cmd reingestType, args ...int32) {
	initConfig()

	i := ingestSystem(ingest.Config{
		IngestFailedTransactions: config.IngestFailedTransactions,
	})
	i.SkipCursorUpdate = true

	logStatus := func(stage string) {
		count := i.Metrics.IngestLedgerTimer.Count()
		rate := i.Metrics.IngestLedgerTimer.RateMean()
		loadMean := time.Duration(i.Metrics.LoadLedgerTimer.Mean())
		ingestMean := time.Duration(i.Metrics.IngestLedgerTimer.Mean())
		clearMean := time.Duration(i.Metrics.ClearLedgerTimer.Mean())
		hlog.WithField("count", count).
			WithField("rate", rate).
			WithField("means", fmt.Sprintf("load: %s clear: %s ingest: %s", loadMean, clearMean, ingestMean)).
			Infof("reingest: %s", stage)
	}

	done := make(chan error, 1)

	// run ingestion in separate goroutine
	go func() {
		var err error
		switch cmd {
		case byAll:
			_, err = i.ReingestAll()

		case bySeq:
			for _, seq := range args {
				err = i.ReingestSingle(seq)
				if err != nil {
					break
				}
			}

		case byRange:
			// should already be checked by the caller
			if len(args) != 2 {
				log.Fatal(`"aurora db reingest range" command requires 2 sequence numbers after "range"`)
			}

			err = reingestRange(i, args[0], args[1])

		case byOutdated:
			_, err = i.ReingestOutdated()
		}

		done <- err
		logStatus("complete")
	}()

	// output metrics
	metrics := time.Tick(2 * time.Second)
	for {
		select {
		case <-metrics:
			logStatus("status")

		case err := <-done:
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)
		}
	}
}

type ledgerRange struct {
	from, to int32
}

func reingestRange(i *ingest.System, from, to int32) error {
	if to < from {
		return errors.New("Invalid range")
	}

	var (
		size    int32 = 10000
		workers int   = 10
	)

	var pool util.WorkersPool
	hlog.Info("Creating work...")
	for current := from; current <= to; current += size {
		lr := ledgerRange{from: current, to: current + size - 1}
		if lr.to > to {
			lr.to = to
		}
		pool.AddWork(lr)
	}

	allJobs := pool.WorkSize()

	pool.SetWorker(func(workerID int, job interface{}) {
		lr, ok := job.(ledgerRange)
		if !ok {
			hlog.Error("job is not a ledgerRange")
			os.Exit(1)
		}

		localLog := hlog.WithFields(hlog.F{
			"id":   workerID,
			"from": lr.from,
			"to":   lr.to,
		})

		localLog.Info("Worker starting range...")

		_, err := i.ReingestRange(lr.from, lr.to)
		if err != nil {
			localLog.WithField("err", err).Error("Worker failed range, work will be processed again")
			// Add the work again
			pool.AddWork(lr)
			return
		}
		localLog.Info("Worker finished range")
	})

	hlog.Infof("Starting %d workers...", workers)
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-done:
				ticker.Stop()
				return
			case <-ticker.C:
			}
			hlog.WithField("progress", float32(allJobs-pool.WorkSize())/float32(allJobs)*100).Info("Work status")
		}
	}()
	pool.Start(workers)
	done <- true
	hlog.Info("Done")
	return nil
}
