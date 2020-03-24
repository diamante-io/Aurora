-- +migrate Up
ALTER TABLE AuthorizedTransaction RENAME TO authorized_transaction;
ALTER TABLE AllowedFI RENAME TO allowed_fi;
ALTER TABLE AllowedUser RENAME TO allowed_user;
ALTER TABLE AuthData RENAME TO auth_data;

-- +migrate Down
ALTER TABLE authorized_transaction RENAME TO AuthorizedTransaction;
ALTER TABLE allowed_fi RENAME TO AllowedFI;
ALTER TABLE allowed_user RENAME TO AllowedUser;
ALTER TABLE auth_data RENAME TO AuthData;