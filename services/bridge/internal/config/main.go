package config

import (
	"errors"
	"net/url"
	"regexp"

	"github.com/diamnet/go/keypair"
	"github.com/diamnet/go/services/internal/bridge-compliance-shared/protocols"
)

// Config contains config params of the bridge server
type Config struct {
	Port              *int              `valid:"required"`
	Aurora           string            `valid:"optional"`
	Compliance        string            `valid:"optional"`
	LogFormat         string            `valid:"optional" toml:"log_format"`
	MACKey            string            `valid:"optional" toml:"mac_key"`
	APIKey            string            `valid:"optional" toml:"api_key"`
	NetworkPassphrase string            `valid:"optional" toml:"network_passphrase"`
	Develop           bool              `valid:"optional"`
	Assets            []protocols.Asset `valid:"optional"`
	Database          *Database         `valid:"optional"`
	Accounts          Accounts          `valid:"optional" toml:"accounts"`
	Callbacks         Callbacks         `valid:"optional" toml:"callbacks"`
}

// Accounts contains values of `accounts` config group
type Accounts struct {
	AuthorizingSeed    string `valid:"optional" toml:"authorizing_seed"`
	BaseSeed           string `valid:"optional" toml:"base_seed"`
	IssuingAccountID   string `valid:"optional" toml:"issuing_account_id"`
	ReceivingAccountID string `valid:"optional" toml:"receiving_account_id"`
}

// Callbacks contains values of `callbacks` config group
type Callbacks struct {
	Receive string `valid:"optional"`
	Error   string `valid:"optional"`
}

// Database contains values of `database` config group
type Database struct {
	Type string `valid:"required"`
	URL  string `valid:"required"`
}

var assetCodeMatch = regexp.MustCompile("^[a-zA-Z0-9]{1,12}$")

// Validate validates config and returns error if any of config values is incorrect
func (c *Config) Validate() (err error) {
	if c.Port == nil {
		err = errors.New("port param is required")
		return
	}

	if c.Aurora == "" {
		err = errors.New("aurora param is required")
		return
	}

	_, err = url.Parse(c.Aurora)
	if err != nil {
		err = errors.New("Cannot parse aurora param")
		return
	}

	if c.NetworkPassphrase == "" {
		err = errors.New("network_passphrase param is required")
		return
	}

	for _, asset := range c.Assets {
		if asset.Issuer == "" {
			if asset.Code != "XLM" {
				err = errors.New("Issuer param is required for " + asset.Code)
				return
			}
		}

		if asset.Issuer != "" {
			_, err = keypair.Parse(asset.Issuer)
			if err != nil {
				err = errors.New("Issuing account is invalid for " + asset.Code)
				return
			}
		}

		matched := assetCodeMatch.MatchString(asset.Code)
		if !matched {
			return errors.New("Invalid asset code: " + asset.Code)
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
	case "":
		// Allow to start gateway server with a single endpoint: /payment
		break
	default:
		err = errors.New("Invalid database.type param")
		return
	}

	if c.Accounts.AuthorizingSeed != "" {
		_, err = keypair.Parse(c.Accounts.AuthorizingSeed)
		if err != nil {
			err = errors.New("accounts.authorizing_seed is invalid")
			return
		}
	}

	if c.Accounts.BaseSeed != "" {
		_, err = keypair.Parse(c.Accounts.BaseSeed)
		if err != nil {
			err = errors.New("accounts.base_seed is invalid")
			return
		}
	}

	if c.Accounts.IssuingAccountID != "" {
		_, err = keypair.Parse(c.Accounts.IssuingAccountID)
		if err != nil {
			err = errors.New("accounts.issuing_account_id is invalid")
			return
		}
	}

	if c.Accounts.ReceivingAccountID != "" {
		_, err = keypair.Parse(c.Accounts.ReceivingAccountID)
		if err != nil {
			err = errors.New("accounts.receiving_account_id is invalid")
			return
		}
	}

	if c.Callbacks.Receive != "" {
		_, err = url.Parse(c.Callbacks.Receive)
		if err != nil {
			err = errors.New("Cannot parse callbacks.receive param")
			return
		}
	}

	if c.Callbacks.Error != "" {
		_, err = url.Parse(c.Callbacks.Error)
		if err != nil {
			err = errors.New("Cannot parse callbacks.error param")
			return
		}
	}

	return
}
