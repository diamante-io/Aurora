-- +migrate Up
CREATE TABLE AuthData (
  id bigserial,
  request_id varchar(255) NOT NULL,
  domain varchar(255) NOT NULL,
  auth_data text NOT NULL,
  
  PRIMARY KEY (id)
);

CREATE UNIQUE INDEX request_id ON AuthData (request_id);

-- +migrate Down
DROP TABLE AuthData;
