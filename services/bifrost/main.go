// Skip this file in Go <1.8 because it's using http.Server.Shutdown
// +build go1.8

package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/facebookgo/inject"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/diamnet/go/clients/aurora"
	"github.com/diamnet/go/services/bifrost/bitcoin"
	"github.com/diamnet/go/services/bifrost/config"
	"github.com/diamnet/go/services/bifrost/database"
	"github.com/diamnet/go/services/bifrost/ethereum"
	"github.com/diamnet/go/services/bifrost/server"
	"github.com/diamnet/go/services/bifrost/sse"
	"github.com/diamnet/go/services/bifrost/diamnet"
	"github.com/diamnet/go/services/bifrost/stress"
	supportConfig "github.com/diamnet/go/support/config"
	"github.com/diamnet/go/support/errors"
	"github.com/diamnet/go/support/log"
)

var rootCmd = &cobra.Command{
	Use:   "bifrost",
	Short: "Bridge server to allow participating in DiamNet based ICOs using Bitcoin and Ethereum",
}

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts backend server",
	Run: func(cmd *cobra.Command, args []string) {
		var (
			cfgPath   = rootCmd.PersistentFlags().Lookup("config").Value.String()
			debugMode = rootCmd.PersistentFlags().Lookup("debug").Changed
		)

		if debugMode {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug mode ON")
		}

		cfg := readConfig(cfgPath)
		server := createServer(cfg, false)
		err := server.Start()
		if err != nil {
			log.WithField("err", err).Error("Error starting the server")
			os.Exit(-1)
		}
	},
}

var stressTestCmd = &cobra.Command{
	Use:   "stress-test",
	Short: "Starts stress test",
	Long: `During stress test bitcoin and ethereum transaction will be generated randomly.
This command will create 3 server.Server's listening on ports 8000-8002.`,
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := rootCmd.PersistentFlags().Lookup("config").Value.String()
		debugMode := rootCmd.PersistentFlags().Lookup("debug").Changed
		usersPerSecond, _ := cmd.PersistentFlags().GetInt("users-per-second")

		if debugMode {
			log.SetLevel(log.DebugLevel)
			log.Debug("Debug mode ON")
		}

		cfg := readConfig(cfgPath)

		bitcoinAccounts := make(chan string)
		bitcoinClient := &stress.RandomBitcoinClient{}
		bitcoinClient.Start(bitcoinAccounts)

		ethereumAccounts := make(chan string)
		ethereumClient := &stress.RandomEthereumClient{}
		ethereumClient.Start(ethereumAccounts)

		db, err := createDatabase(cfg.Database.DSN)
		if err != nil {
			log.WithField("err", err).Error("Error connecting to database")
			os.Exit(-1)
		}

		err = db.ResetBlockCounters()
		if err != nil {
			log.WithField("err", err).Error("Error reseting counters")
			os.Exit(-1)
		}

		// Start servers
		const numServers = 3
		signers := []string{
			// GBQYGXC4AZDL7PPL2H274LYA6YV7OL4IRPWCYMBYCA5FAO45WMTNKGOD
			"SBX76SCADD2SBIL6M2T62BR4GELMJPZV2MFHIQX24IBVOTIT6DGNAR3D",
			// GAUDK66OCTKQB737ZNRD2ILB5ZGIZOKMWT3T5TDWVIN7ANVY3RD5DXF3
			"SCGJ6JRFMHWYTGVBPFXCBMMBENZM433M3JNZDRSI5PZ2DXGJLYDWR4CR",
			// GDFR6QNVBUK32PGTAV3HATV3GDT7LF2SGVVLD2TOS4TCAD2ANSOH2MCW
			"SAZUY2XGSILNMBLQSVMDGCCTSZNOB2EXHSFFRFJJ3GKRZIW3FTIMJYV7",
		}
		ports := []int{8000, 8001, 8002}
		for i := 0; i < numServers; i++ {
			go func(i int) {
				cfg.Port = ports[i]
				cfg.DiamNet.SignerSecretKey = signers[i]
				server := createServer(cfg, true)
				// Replace clients in listeners with random transactions generators
				server.BitcoinListener.Client = bitcoinClient
				server.EthereumListener.Client = ethereumClient
				err := server.Start()
				if err != nil {
					log.WithField("err", err).Error("Error starting the server")
					os.Exit(-1)
				}
			}(i)
		}

		// Wait for servers to start. We could wait in a more sophisticated way but this
		// is just a test code.
		time.Sleep(2 * time.Second)

		accounts := make(chan server.GenerateAddressResponse)
		users := stress.Users{
			Aurora: &aurora.Client{
				URL: cfg.DiamNet.Aurora,
				HTTP: &http.Client{
					Timeout: 60 * time.Second,
				},
				AppName: "bifrost",
			},
			NetworkPassphrase: cfg.DiamNet.NetworkPassphrase,
			UsersPerSecond:    usersPerSecond,
			BifrostPorts:      ports,
			IssuerPublicKey:   cfg.DiamNet.IssuerPublicKey,
		}
		go users.Start(accounts)
		for {
			account := <-accounts
			switch account.Chain {
			case string(database.ChainBitcoin):
				bitcoinAccounts <- account.Address
			case string(database.ChainEthereum):
				ethereumAccounts <- account.Address
			default:
				panic("Unknown chain: " + account.Chain)
			}
		}
	},
}

var checkKeysCmd = &cobra.Command{
	Use:   "check-keys",
	Short: "Displays a few public keys derived using master public keys",
	Run: func(cmd *cobra.Command, args []string) {
		cfgPath := rootCmd.PersistentFlags().Lookup("config").Value.String()
		start, _ := cmd.PersistentFlags().GetUint32("start")
		count, _ := cmd.PersistentFlags().GetUint32("count")
		cfg := readConfig(cfgPath)

		fmt.Println("MAKE SURE YOU HAVE PRIVATE KEYS TO CORRESPONDING ADDRESSES:")

		fmt.Println("Bitcoin MainNet:")
		if cfg.Bitcoin != nil && cfg.Bitcoin.MasterPublicKey != "" {
			bitcoinAddressGenerator, err := bitcoin.NewAddressGenerator(cfg.Bitcoin.MasterPublicKey, &chaincfg.MainNetParams)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}

			for i := uint32(start); i < start+count; i++ {
				address, err := bitcoinAddressGenerator.Generate(i)
				if err != nil {
					fmt.Println("Error generating address", i)
					continue
				}
				fmt.Printf("%d %s\n", i, address)
			}
		} else {
			fmt.Println("No master key set...")
		}

		fmt.Println("Ethereum:")
		if cfg.Ethereum != nil && cfg.Ethereum.MasterPublicKey != "" {
			ethereumAddressGenerator, err := ethereum.NewAddressGenerator(cfg.Ethereum.MasterPublicKey)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}

			for i := uint32(start); i < start+count; i++ {
				address, err := ethereumAddressGenerator.Generate(i)
				if err != nil {
					fmt.Println("Error generating address", i)
					continue
				}
				fmt.Printf("%d %s\n", i, address)
			}
		} else {
			fmt.Println("No master key set...")
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.2")
	},
}

func init() {
	// TODO I think these should be default in diamnet/go:
	log.SetLevel(log.InfoLevel)
	log.DefaultLogger.Logger.Formatter.(*logrus.TextFormatter).FullTimestamp = true

	rootCmd.PersistentFlags().Bool("debug", false, "debug mode")
	rootCmd.PersistentFlags().StringP("config", "c", "bifrost.cfg", "config file path")

	rootCmd.AddCommand(checkKeysCmd)
	rootCmd.AddCommand(serverCmd)
	rootCmd.AddCommand(stressTestCmd)
	rootCmd.AddCommand(versionCmd)

	stressTestCmd.PersistentFlags().IntP("users-per-second", "u", 2, "users per second")

	checkKeysCmd.PersistentFlags().Uint32P("start", "s", 0, "starting address index")
	checkKeysCmd.PersistentFlags().Uint32P("count", "l", 10, "how many addresses generate")
}

func main() {
	rootCmd.Execute()
}

func readConfig(cfgPath string) config.Config {
	var cfg config.Config

	err := supportConfig.Read(cfgPath, &cfg)
	if err != nil {
		switch cause := errors.Cause(err).(type) {
		case *supportConfig.InvalidConfigError:
			log.Error("config file: ", cause)
		default:
			log.Error(err)
		}
		os.Exit(-1)
	}
	if cfg.AccessControlAllowOriginHeader == "" {
		cfg.AccessControlAllowOriginHeader = "*"
	}

	return cfg
}

// createDatabase opens a DB connection and imports schema if DB is empty.
func createDatabase(dsn string) (*database.PostgresDatabase, error) {
	db := &database.PostgresDatabase{}
	err := db.Open(dsn)
	if err != nil {
		return nil, err
	}

	currentSchemaVersion, err := db.GetSchemaVersion()
	if err != nil {
		return nil, err
	}

	if currentSchemaVersion == 0 {
		// DB clean, import
		err := db.Import()
		if err != nil {
			return nil, errors.Wrap(err, "Error importing DB schema")
		}
	} else if currentSchemaVersion != database.SchemaVersion {
		// Schema version invalid
		return nil, errors.New("Schema version is invalid. Please create an empty DB and start Bifrost again.")
	}

	return db, nil
}

func createServer(cfg config.Config, stressTest bool) *server.Server {
	var g inject.Graph

	db, err := createDatabase(cfg.Database.DSN)
	if err != nil {
		log.WithField("err", err).Error("Error connecting to database")
		os.Exit(-1)
	}

	server := &server.Server{
		SignerPublicKey: cfg.SignerPublicKey(),
	}

	bitcoinClient := &rpcclient.Client{}
	bitcoinListener := &bitcoin.Listener{}
	bitcoinAddressGenerator := &bitcoin.AddressGenerator{}

	ethereumClient := &ethclient.Client{}
	ethereumListener := &ethereum.Listener{}
	ethereumAddressGenerator := &ethereum.AddressGenerator{}

	if !stressTest {
		// Configure real clients
		if cfg.Bitcoin != nil {
			connConfig := &rpcclient.ConnConfig{
				Host:         cfg.Bitcoin.RpcServer,
				User:         cfg.Bitcoin.RpcUser,
				Pass:         cfg.Bitcoin.RpcPass,
				HTTPPostMode: true,
				DisableTLS:   true,
			}
			bitcoinClient, err = rpcclient.New(connConfig, nil)
			if err != nil {
				log.WithField("err", err).Error("Error connecting to bitcoin-core")
				os.Exit(-1)
			}

			bitcoinListener.Enabled = true
			bitcoinListener.Testnet = cfg.Bitcoin.Testnet
			server.MinimumValueBtc = cfg.Bitcoin.MinimumValueBtc

			var chainParams *chaincfg.Params
			if cfg.Bitcoin.Testnet {
				chainParams = &chaincfg.TestNet3Params
			} else {
				chainParams = &chaincfg.MainNetParams
			}
			bitcoinAddressGenerator, err = bitcoin.NewAddressGenerator(cfg.Bitcoin.MasterPublicKey, chainParams)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
		}

		if cfg.Ethereum != nil {
			ethereumClient, err = ethclient.Dial("http://" + cfg.Ethereum.RpcServer)
			if err != nil {
				log.WithField("err", err).Error("Error connecting to geth")
				os.Exit(-1)
			}

			ethereumListener.Enabled = true
			ethereumListener.NetworkID = cfg.Ethereum.NetworkID
			server.MinimumValueEth = cfg.Ethereum.MinimumValueEth

			ethereumAddressGenerator, err = ethereum.NewAddressGenerator(cfg.Ethereum.MasterPublicKey)
			if err != nil {
				log.Error(err)
				os.Exit(-1)
			}
		}
	} else {
		bitcoinListener.Enabled = true
		bitcoinListener.Testnet = true
		bitcoinAddressGenerator, err = bitcoin.NewAddressGenerator(cfg.Bitcoin.MasterPublicKey, &chaincfg.TestNet3Params)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}

		ethereumListener.Enabled = true
		ethereumListener.NetworkID = "3"
		ethereumAddressGenerator, err = ethereum.NewAddressGenerator(cfg.Ethereum.MasterPublicKey)
		if err != nil {
			log.Error(err)
			os.Exit(-1)
		}
	}

	diamnetAccountConfigurator := &diamnet.AccountConfigurator{
		NetworkPassphrase:     cfg.DiamNet.NetworkPassphrase,
		IssuerPublicKey:       cfg.DiamNet.IssuerPublicKey,
		DistributionPublicKey: cfg.DiamNet.DistributionPublicKey,
		SignerSecretKey:       cfg.DiamNet.SignerSecretKey,
		NeedsAuthorize:        cfg.DiamNet.NeedsAuthorize,
		TokenAssetCode:        cfg.DiamNet.TokenAssetCode,
		StartingBalance:       cfg.DiamNet.StartingBalance,
		LockUnixTimestamp:     cfg.DiamNet.LockUnixTimestamp,
	}

	if cfg.DiamNet.StartingBalance == "" {
		diamnetAccountConfigurator.StartingBalance = "2.1"
	}

	if cfg.Bitcoin != nil {
		diamnetAccountConfigurator.TokenPriceBTC = cfg.Bitcoin.TokenPrice
	}

	if cfg.Ethereum != nil {
		diamnetAccountConfigurator.TokenPriceETH = cfg.Ethereum.TokenPrice
	}

	auroraClient := &aurora.Client{
		URL: cfg.DiamNet.Aurora,
		HTTP: &http.Client{
			Timeout: 20 * time.Second,
		},
		AppName: "bifrost",
	}

	sseServer := &sse.Server{}

	err = g.Provide(
		&inject.Object{Value: bitcoinAddressGenerator},
		&inject.Object{Value: bitcoinClient},
		&inject.Object{Value: bitcoinListener},
		&inject.Object{Value: &cfg},
		&inject.Object{Value: db},
		&inject.Object{Value: ethereumAddressGenerator},
		&inject.Object{Value: ethereumClient},
		&inject.Object{Value: ethereumListener},
		&inject.Object{Value: auroraClient},
		&inject.Object{Value: server},
		&inject.Object{Value: sseServer},
		&inject.Object{Value: diamnetAccountConfigurator},
	)
	if err != nil {
		log.WithField("err", err).Error("Error providing objects to injector")
		os.Exit(-1)
	}

	if err := g.Populate(); err != nil {
		log.WithField("err", err).Error("Error injecting objects")
		os.Exit(-1)
	}

	return server
}
