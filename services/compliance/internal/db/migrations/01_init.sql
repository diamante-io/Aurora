-- +migrate Up
CREATE TABLE AuthorizedTransaction (
  id bigserial,
  transaction_id varchar(64) NOT NULL,
  memo varchar(64) NOT NULL,
  transaction_xdr text NOT NULL,
  authorized_at timestamp NOT NULL,
  data text NOT NULL,
  
  PRIMARY KEY (id)
);

CREATE TABLE AllowedFI (
  id bigserial,
  name varchar(255) NOT NULL,
  domain varchar(255) NOT NULL,
  public_key char(56) NOT NULL,
  allowed_at timestamp NOT NULL,
  PRIMARY KEY (id)

) ;

CREATE UNIQUE INDEX afi_by_domain ON AllowedFI (domain);
CREATE UNIQUE INDEX afi_by_public_key ON AllowedFI (public_key);

CREATE TABLE AllowedUser (
  id bigserial,
  fi_name varchar(255) NOT NULL,
  fi_domain varchar(255) NOT NULL,
  fi_public_key char(56) NOT NULL,
  user_id varchar(255) NOT NULL,
  allowed_at timestamp NOT NULL,
  PRIMARY KEY (id)
);

CREATE UNIQUE INDEX au_by_fi_public_key_user_id ON AllowedUser (fi_public_key, user_id);


-- +migrate Down
DROP TABLE AuthorizedTransaction;
DROP TABLE AllowedFI;
DROP TABLE AllowedUser;
