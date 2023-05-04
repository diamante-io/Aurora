package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/matryer/try.v1"
)

const stelExURL = "https://api.diamnet.expert/explorer/public/xlm-price"

const ratesURL = "https://openexchangerates.org/api/latest.json"

type cachedPrice struct {
	price   float64
	updated time.Time
}

func mustCreateXlmPriceRequest() *http.Request {
	numAttempts := 10
	var req *http.Request
	err := try.Do(func(attempt int) (bool, error) {
		var err error
		req, err = createXlmPriceRequest()
		if err != nil {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		return attempt < numAttempts, err
	})
	if err != nil {
		// TODO: Add a fallback price API.
		log.Fatal(err)
	}
	return req
}

func createXlmPriceRequest() (*http.Request, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", stelExURL, nil)
	if err != nil {
		return nil, err
	}

	// TODO: Eliminate dependency on dotenv before monorepo conversion.
	authKey := os.Getenv("DIAMNET_EXPERT_AUTH_KEY")
	authVal := os.Getenv("DIAMNET_EXPERT_AUTH_VAL")
	req.Header.Add(authKey, authVal)

	return req, nil
}

func getLatestXlmPrice(req *http.Request) (float64, error) {
	body, err := getPriceResponse(req)
	if err != nil {
		return 0.0, fmt.Errorf("got error from diamnet expert price api: %s", err)
	}
	return parseDiamnetExpertLatestPrice(body)
}

func getXlmPriceHistory(req *http.Request) ([]xlmPrice, error) {
	body, err := getPriceResponse(req)
	if err != nil {
		return []xlmPrice{}, fmt.Errorf("got error from diamnet expert price api: %s", err)
	}
	return parseDiamnetExpertPriceHistory(body)
}

func getPriceResponse(req *http.Request) (string, error) {
	client := &http.Client{}

	numAttempts := 10
	var resp *http.Response
	err := try.Do(func(attempt int) (bool, error) {

		var err error
		resp, err = client.Do(req)
		if err != nil {
			return attempt < numAttempts, err
		}

		if resp.StatusCode != http.StatusOK {
			time.Sleep(time.Duration(attempt) * time.Second)
			err = fmt.Errorf("got status code %d", resp.StatusCode)
		}

		return attempt < numAttempts, err
	})

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	body := string(bodyBytes)
	return body, nil
}

func parseDiamnetExpertPriceHistory(body string) ([]xlmPrice, error) {
	// The Diamnet Expert response has expected format: [[timestamp1,price1], [timestamp2,price2], ...]
	// with the most recent timestamp and price first. We split that array to get strings of only "timestamp,price".
	// We then split each of those strings and define a struct containing the timestamp and price.
	if len(body) < 5 {
		return []xlmPrice{}, fmt.Errorf("got ill-formed response body from diamnet expert")
	}

	body = body[2 : len(body)-2]
	timePriceStrs := strings.Split(body, "],[")

	var xlmPrices []xlmPrice
	for _, timePriceStr := range timePriceStrs {
		timePrice := strings.Split(timePriceStr, ",")
		if len(timePrice) != 2 {
			return []xlmPrice{}, fmt.Errorf("got ill-formed time/price from diamnet expert")
		}

		ts, err := strconv.ParseInt(timePrice[0], 10, 64)
		if err != nil {
			return []xlmPrice{}, err
		}

		p, err := strconv.ParseFloat(timePrice[1], 64)
		if err != nil {
			return []xlmPrice{}, err
		}

		newXlmPrice := xlmPrice{
			timestamp: ts,
			price:     p,
		}
		xlmPrices = append(xlmPrices, newXlmPrice)
	}
	return xlmPrices, nil
}

func parseDiamnetExpertLatestPrice(body string) (float64, error) {
	// The Diamnet Expert response has expected format: [[timestamp1,price1], [timestamp2,price2], ...]
	// with the most recent timestamp and price first.
	// We then split the remainder by ",".
	// The first element will be the most recent timestamp, and the second will be the latest price.
	// We format and return the most recent price.
	lists := strings.Split(body, ",")
	if len(lists) < 2 {
		return 0.0, fmt.Errorf("mis-formed response from diamnet expert")
	}

	rawPriceStr := lists[1]
	if len(rawPriceStr) < 2 {
		return 0.0, fmt.Errorf("mis-formed price from diamnet expert")
	}

	priceStr := rawPriceStr[:len(rawPriceStr)-1]
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return 0.0, err
	}

	return price, nil
}

func mustCreateAssetPriceRequest() *http.Request {
	numAttempts := 10
	var req *http.Request
	err := try.Do(func(attempt int) (bool, error) {
		var err error
		req, err = createAssetPriceRequest()
		if err != nil {
			time.Sleep(time.Duration(attempt) * time.Second)
		}
		return attempt < numAttempts, err
	})
	if err != nil {
		// TODO: Add a fallback price API.
		log.Fatal(err)
	}
	return req
}

func createAssetPriceRequest() (*http.Request, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", ratesURL, nil)
	if err != nil {
		return nil, err
	}

	// TODO: Eliminate dependency on dotenv before monorepo conversion.
	apiKey := os.Getenv("RATES_API_KEY")
	apiVal := os.Getenv("RATES_API_VAL")
	req.Header.Add(apiKey, apiVal)
	return req, nil
}

func getAssetUSDPrice(body, currency string) (float64, error) {
	// The real asset price for USD will be 1 USD
	if currency == "USD" {
		return 1.0, nil
	} else if currency == "" {
		return 0.0, nil
	}

	// we expect the body to contain a JSON response from the OpenExchangeRates API,
	// including a "rates" field which maps currency code to USD rate.
	// e.g., "USD": 1.0, "BRL": 5.2, etc.
	m := make(map[string]interface{})
	json.Unmarshal([]byte(body), &m)

	rates := make(map[string]interface{})
	var ok bool
	if rates, ok = m["rates"].(map[string]interface{}); !ok {
		return 0.0, fmt.Errorf("could not get rates from api response")
	}

	var rate float64
	if rate, ok = rates[currency].(float64); !ok {
		return 0.0, fmt.Errorf("could not get rate for %s", currency)
	}

	return rate, nil
}

func updateAssetUsdPrice(currency string) (float64, error) {
	assetReq, err := createAssetPriceRequest()
	if err != nil {
		return 0.0, fmt.Errorf("could not create asset price request: %s", err)
	}

	assetMapStr, err := getPriceResponse(assetReq)
	if err != nil {
		return 0.0, fmt.Errorf("could not get asset price response from external api: %s", err)
	}

	assetUsdPrice, err := getAssetUSDPrice(assetMapStr, currency)
	if err != nil {
		return 0.0, fmt.Errorf("could not parse asset price response from external api: %s", err)
	}

	return assetUsdPrice, nil
}

func createPriceCache(pairs []prometheusWatchedTP) map[string]cachedPrice {
	pc := make(map[string]cachedPrice)
	t := time.Now().Add(-2 * time.Hour)
	for _, p := range pairs {
		pc[p.TradePair.BuyingAsset.Currency] = cachedPrice{
			price:   0.0,
			updated: t,
		}
	}
	return pc
}
