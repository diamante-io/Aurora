-- +migrate Up
ALTER TABLE SentTransaction ADD payment_id VARCHAR(255) NULL DEFAULT NULL;
ALTER TABLE SentTransaction ADD CONSTRAINT payment_id_unique UNIQUE (payment_id);

-- +migrate Down
ALTER TABLE SentTransaction DROP payment_id;
