-- +migrate Up
ALTER TABLE ReceivedPayment RENAME TO received_payment;
ALTER TABLE SentTransaction RENAME TO sent_transaction;

-- +migrate Down
ALTER TABLE received_payment RENAME TO ReceivedPayment;
ALTER TABLE sent_transaction RENAME TO SentTransaction;
