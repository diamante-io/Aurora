package database

import (
	"fmt"
)

var schema = fmt.Sprintf(`CREATE TYPE chain AS ENUM ('bitcoin', 'ethereum');

CREATE TABLE address_association (
  chain chain NOT NULL,
  address_index bigint NOT NULL,
  /* bitcoin 34 characters */
  /* ethereum 42 characters */
  address varchar(42) NOT NULL UNIQUE,
  hcnet_public_key varchar(56) NOT NULL UNIQUE,
  created_at timestamp NOT NULL,
  PRIMARY KEY (chain, address_index, address, hcnet_public_key),
  CONSTRAINT valid_address_index CHECK (address_index >= 0)
);

CREATE TABLE key_value_store (
  key varchar(255) NOT NULL,
  value varchar(255) NOT NULL,
  PRIMARY KEY (key)
);

INSERT INTO key_value_store (key, value) VALUES ('schema_version', '%d');

INSERT INTO key_value_store (key, value) VALUES ('ethereum_address_index', '0');
INSERT INTO key_value_store (key, value) VALUES ('ethereum_last_block', '0');

INSERT INTO key_value_store (key, value) VALUES ('bitcoin_address_index', '0');
INSERT INTO key_value_store (key, value) VALUES ('bitcoin_last_block', '0');

CREATE TABLE processed_transaction (
  chain chain NOT NULL,
  /* Ethereum: "0x"+hash (so 64+2) */
  transaction_id varchar(66) NOT NULL,
  /* bitcoin 34 characters */
  /* ethereum 42 characters */
  receiving_address varchar(42) NOT NULL,
  created_at timestamp NOT NULL,
  PRIMARY KEY (chain, transaction_id)
);

/* If using DB storage for the queue not AWS FIFO */
CREATE TABLE transactions_queue (
  id bigserial,
  /* Ethereum: "0x"+hash (so 64+2) */
  transaction_id varchar(66) NOT NULL,
  asset_code varchar(3) NOT NULL,
  /* Amount in the base unit of currency (BTC or ETH). */
  /* ethereum: 100000000 in year 2128 + 7 decimal precision in HcNet + dot */
  /* bitcoin:   21000000              + 7 decimal precision in HcNet + dot */
  amount varchar(20) NOT NULL,
  hcnet_public_key varchar(56) NOT NULL,
  pooled boolean NOT NULL DEFAULT false,
  PRIMARY KEY (id),
  UNIQUE (transaction_id, asset_code),
  CONSTRAINT valid_asset_code CHECK (char_length(asset_code) = 3),
  CONSTRAINT valid_hcnet_public_key CHECK (char_length(hcnet_public_key) = 56)
);

CREATE TYPE event AS ENUM ('transaction_received', 'account_created', 'exchanged', 'exchanged_timelocked');

CREATE TABLE broadcasted_event (
  id bigserial,
  /* bitcoin 34 characters */
  /* ethereum 42 characters */
  address varchar(42) NOT NULL,
  event event NOT NULL,
  data text NOT NULL,
  PRIMARY KEY (id),
  UNIQUE (address, event)
);

CREATE TABLE recovery_transaction (
  source varchar(56) NOT NULL,
  envelope_xdr text NOT NULL
);

CREATE INDEX source_index ON recovery_transaction (source);`, SchemaVersion)
