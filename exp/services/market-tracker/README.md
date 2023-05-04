# Diamnet Market Tracker

The Diamnet Market Tracker allows you to monitor the spreads of any desired asset pairs, and makes them available for Prometheus to scrape.
To use this project, you will have to define the following in a `.env`:
- Custom `config.json` listing the asset pairs for monitoring. The format is displayed in `config_sample.json`
- Environment variables: `DIAMNET_EXPERT_AUTH_KEY` and `DIAMNET_EXPERT_AUTH_VAL`, the authentication header for Diamnet Expert; `RATES_API_KEY` and `RATES_API_VAL`, the key-value pair for the OpenExchangeRates API. Note that the exact format of these variables may change, as we finalize internal deployment.

## Running the project

This project was built using Go (this repository is officially supported on the last two releases of Go) and [Go Modules](https://blog.golang.org/using-go-modules)

1. From the monorepo root, navigate to the project: `cd exp/services/market-tracker`
2. Create a `config.json` file with the asset pairs to monitor and the refresh interval. A sample file is checked in at `config_sample.json`
3. Build the project: `go build .`
4. Run the project `./market-tracker`
5. Open `http://127.0.01:2112/metrics` and you should be able to view the metrics. This is the endpoint Prometheus should scrape.
