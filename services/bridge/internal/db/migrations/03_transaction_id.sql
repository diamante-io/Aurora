-- +migrate Up
ALTER TABLE ReceivedPayment ADD transaction_id VARCHAR(64) DEFAULT 'N/A';

-- +migrate Down
ALTER TABLE ReceivedPayment DROP transaction_id;
