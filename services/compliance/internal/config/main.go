package config

import (
	"errors"
	"net/url"

	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/support/config"
)

// Config contains config params of the compliance server
type Config struct {
	ExternalPort      *int          `valid:"required" toml:"external_port"`
	InternalPort      *int          `valid:"required" toml:"internal_port"`
	LogFormat         string        `valid:"optional" toml:"log_format"`
	NeedsAuth         bool          `valid:"optional" toml:"needs_auth"`
	NetworkPassphrase string        `valid:"required" toml:"network_passphrase"`
	Database          Database      `valid:"required"`
	Keys              Keys          `valid:"required" toml:"keys"`
	Callbacks         Callbacks     `valid:"optional" toml:"callbacks"`
	TLS               *config.TLS   `valid:"optional"`
	TxStatusAuth      *TxStatusAuth `valid:"optional" toml:"tx_status_auth"`
}

type TxStatusAuth struct {
	Username string `valid:"required" toml:"username"`
	Password string `valid:"required" toml:"password"`
}

// Keys contains values of `keys` config group
type Keys struct {
	SigningSeed string `valid:"required" toml:"signing_seed"`
}

// Callbacks contains values of `callbacks` config group
type Callbacks struct {
	Sanctions string `valid:"optional"`
	AskUser   string `valid:"optional" toml:"ask_user"`
	FetchInfo string `valid:"optional" toml:"fetch_info"`
	TxStatus  string `valid:"optional" toml:"tx_status"`
}

// Database contains values of `database` config group
type Database struct {
	Type string `valid:"required"`
	URL  string `valid:"required"`
}

// Validate validates config and returns error if any of config values is incorrect
func (c *Config) Validate() (err error) {
	if c.ExternalPort == nil {
		err = errors.New("external_port param is required")
		return
	}

	if c.InternalPort == nil {
		err = errors.New("internal_port param is required")
		return
	}

	if c.NetworkPassphrase == "" {
		err = errors.New("network_passphrase param is required")
		return
	}

	if c.Keys.SigningSeed == "" {
		err = errors.New("keys.signing_seed and keys.encryption_key params are required")
		return
	}

	if c.Keys.SigningSeed != "" {
		_, err = keypair.Parse(c.Keys.SigningSeed)
		if err != nil {
			err = errors.New("keys.signing_seed is invalid")
			return
		}
	}

	var dbURL *url.URL
	dbURL, err = url.Parse(c.Database.URL)
	if err != nil {
		err = errors.New("Cannot parse database.url param")
		return
	}

	switch c.Database.Type {
	case "mysql":
		// Add `parseTime=true` param to mysql url
		query := dbURL.Query()
		query.Set("parseTime", "true")
		dbURL.RawQuery = query.Encode()
		c.Database.URL = dbURL.String()
	case "postgres":
		break
	default:
		err = errors.New("Invalid database.type param")
		return
	}

	if c.Callbacks.Sanctions != "" {
		_, err = url.Parse(c.Callbacks.Sanctions)
		if err != nil {
			err = errors.New("Cannot parse callbacks.sanctions param")
			return
		}
	}

	if c.Callbacks.TxStatus != "" {
		_, err = url.Parse(c.Callbacks.TxStatus)
		if err != nil {
			err = errors.New("Cannot parse callbacks.tx_status param")
			return
		}
	}

	if c.Callbacks.AskUser != "" {
		_, err = url.Parse(c.Callbacks.AskUser)
		if err != nil {
			err = errors.New("Cannot parse callbacks.ask_user param")
			return
		}
	}

	if c.Callbacks.FetchInfo != "" {
		_, err = url.Parse(c.Callbacks.FetchInfo)
		if err != nil {
			err = errors.New("Cannot parse callbacks.fetch_info param")
			return
		}
	}

	return
}
