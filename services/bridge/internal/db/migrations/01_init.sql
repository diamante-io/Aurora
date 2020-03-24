-- +migrate Up
CREATE TABLE ReceivedPayment (
  id bigserial,
  operation_id varchar(255) UNIQUE NOT NULL,
  processed_at timestamp NOT NULL,
  paging_token varchar(255) NOT NULL,
  status varchar(255) NOT NULL,
  PRIMARY KEY (id)
);

CREATE TABLE SentTransaction (
  id serial,
  transaction_id varchar(64) NOT NULL, 
  status varchar(10) NOT NULL,
  source varchar(56) NOT NULL,
  submitted_at timestamp NOT NULL,
  succeeded_at timestamp DEFAULT NULL,
  ledger bigint DEFAULT NULL,
  envelope_xdr text NOT NULL,
  result_xdr varchar(255) DEFAULT NULL,
  PRIMARY KEY (id)
);

-- +migrate Down
DROP TABLE ReceivedPayment;
DROP TABLE SentTransaction;
