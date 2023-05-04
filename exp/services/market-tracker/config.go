package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	hClient "github.com/diamnet/go/clients/auroraclient"
)

// Asset represents an asset on the Diamnet network
type Asset struct {
	ProtocolAssetType hClient.AssetType
	AssetType         string `json:"type"`
	Code              string `json:"code"`
	IssuerAddress     string `json:"issuerAddress"`
	IssuerName        string `json:"issuerName"`
	Currency          string `json:"currency"`
}

func (a Asset) String() string {
	if a.IssuerName != "" {
		return fmt.Sprintf("%s:%s", a.Code, a.IssuerName)
	}
	return fmt.Sprintf("%s:%s", a.Code, a.IssuerAddress)
}

// TradePair represents a trading pair on SDEX
type TradePair struct {
	BuyingAsset  Asset `json:"buyingAsset"`
	SellingAsset Asset `json:"sellingAsset"`
}

func (tp TradePair) String() string {
	return fmt.Sprintf("%s / %s", tp.BuyingAsset, tp.SellingAsset)
}

// Config represents the overall config of the application
type Config struct {
	TradePairs           []TradePair `json:"tradePairs"`
	CheckIntervalSeconds int64       `json:"checkIntervalSeconds"`
}

func computeAssetType(a *Asset) (err error) {
	switch a.AssetType {
	case "AssetType4":
		a.ProtocolAssetType = hClient.AssetType4
	case "AssetType12":
		a.ProtocolAssetType = hClient.AssetType12
	case "AssetTypeNative":
		a.ProtocolAssetType = hClient.AssetTypeNative
	default:
		err = errors.New("unrecognized asset type")
	}
	return
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func loadConfig() Config {
	configFile, err := os.Open("config_sample.json")
	check(err)

	defer configFile.Close()

	byteValue, _ := ioutil.ReadAll(configFile)

	var config Config
	err = json.Unmarshal(byteValue, &config)
	check(err)

	for n := range config.TradePairs {
		err = computeAssetType(&config.TradePairs[n].BuyingAsset)
		check(err)

		computeAssetType(&config.TradePairs[n].SellingAsset)
		check(err)
	}

	return config
}
