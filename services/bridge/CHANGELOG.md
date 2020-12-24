# Changelog

As this project is pre 1.0, breaking changes may happen for minor version bumps. A breaking change will get clearly notified in this log.

## Unreleased

* Fixed path-payment operation in `/payment`

## 0.0.32

* Bridge server now uses the new Go SDK.
* Unit tests added.

## Breaking changes

* MySQL is no longer supported. To migrate your data to postgresql use any of the tools provided [here](https://wiki.postgresql.org/wiki/Converting_from_other_Databases_to_PostgreSQL#MySQL).

* Requests to `/builder` for `AllowTrust` operations must specify the source account for the operation.


## Changes
* Payload MAC authentication uses `X-Payload-Mac` header (old `X_PAYLOAD_MAC` header is still provided for backward compatibility, but it is deprecated and will be removed in future versions).

## 0.0.31

### Breaking changes
* `id` parameter is now required when sending payments using Compliance Protocol.

### Changes
* `nonce` value does not change when repeating Auth Request after receiving `pending` status.
* Support for sending XLM using Compliance Protocol.
* Fix for #109

Please migrate your `compliance` DB before running a new version using: `compliance --migrate-db`.

## 0.0.30

* Support for "Forward federation" destinations.

## 0.0.29

* Improved transaction submission code:
  * High rate transaction submission using `/payment` endpoint should work better.
  * Added `id` parameter to `/payment` request: payments with `id` set, when resubmitted, are using previously created transaction envelope stored in a DB instead of recreating a transaction with a new sequence number. This can prevent accidental double-spends.
* Fix for a bug in `/builder` endpoint: sequence number is now incremented when loaded from Aurora server (https://github.com/diamnet/bridge-server/issues/86).
* Payment listener is now also sending `account_merge` operations and, for each operation, a new parameter: `transaction_id`.
* Updated `github.com/BurntSushi/toml` dependency.

Read `README` file for more information about new features.

Please migrate your `bridge` DB before running a new version using: `bridge --migrate-db`

## 0.0.28

* Added error messages to Compliance protocol ([SEP-0003](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0003.md))

## 0.0.27

* Admin Panel (`/admin` endpoint in `bridge` server).
* `/tx_status` endpoint [More info](https://github.com/diamnet/diamnet-protocol/blob/master/ecosystem/sep-0001.md).
* Sequence number in automatically loaded if it's not set in `/builder`.
* Fixed log levels in `PaymentListener` (#73).
* Fixed `AllowedFI` table name under Windows (#72).
* New `-v` parameter to print app version.

## 0.0.26

* Fix log level in `PaymentListener`.

## 0.0.25

* [XLM (lumen)](https://www.diamnet.org/lumens/) payments can be now used in `PaymentListener`.
* Fixed a loop in `PaymentListener` occurring when multiple payments fail.

## 0.0.24

* Better responses.
* Use `http.Client` with `Timeout`.

## 0.0.23

* Fix a bug in `protocols.Asset.String`. Add more_info field to `invalid_parameter` errors.

## 0.0.22

* Ability to reprocess received payments.

## 0.0.21

* Add asset issuer to receive callback

## 0.0.20

* Update `github.com/diamnet/go` dependency.

## 0.0.19

* Fix account ID destinations in /payment

## 0.0.18

* Fixed `-config` param in bridge.

## 0.0.17

* Added `-config` param to use custom config file
* Removed unused `EncryptionKey` config param
* Added use_compliance parameter in `/payment` to force using compliance protocol.

## 0.0.16

* New version of compliance protocol.

## 0.0.15

* Change diamnet.toml location to new standard.

## 0.0.14

* Bug fixes for postgres

## 0.0.13

* Add `mac_key` configuration

## 0.0.12

* Fix `inject` in compliance server.

## 0.0.11

* `/create-keypair` endpoint,
* Sending routing information in receive callback.

## 0.0.10

* Send only relevant data to compliance callbacks (#17).
* `hooks` are now called `callbacks` in `bridge` server.

## 0.0.9

* Transaction builder (#14)

## 0.0.8

* [Compliance protocol](https://www.diamnet.org/developers/learn/integration-guides/compliance-protocol.html) support.
* Saving and reading memo preimage.
* This repo will now contain two apps: `bridge` (for building, submitting and monitoring transactions) and `compliance` (for Compliance protocol). Both are built in a single build process. Each app has it's own README file.
* Dependency injection is now done using [facebookgo/inject](https://godoc.org/github.com/facebookgo/inject).
* Handling and validation of requests and responses is now done in `protocols` package. This package contains methods for transforming `url.Values` from/to request structs and for marshalling responses. It also contains common errors (missing/invalid fields, internal server error, etc.) and all protocol-specific error responses. It also includes diamnet.toml and federation resolving.
* New `net` and `server` packages that contain some helper network connected functions and structs.
* Improvements to `db` package.

## 0.0.7

* Add path payments,
* Change `config.toml` file structure,
* Partial implementation of Compliance Protocol.

## 0.0.6

* When there are no `ReceivePayment`s in database, payment listener will start with cursor `now`.
* Fix a bug in `db.Repository.GetLastCursorValue()`.

## 0.0.5

* Add `MEMO_HASH` support.

## 0.0.4

* Fixed bugs connected with running server using `postgres` DB (full refactoring of `db` package),
* Fixed starting a minimum server with a single endpoint: `/payment`.

## 0.0.3

* Send `create_account` operation in `/payment` if account does not exist.
* Fixed major bug in `PaymentListener`.
* Sending to DiamNet address with memo in `/send`.
* Standardized responses.
* Updated README file.

## 0.0.2

* Added `/payment` endpoint.
* Now it's possible to start a server with parameter that are not required. Minimum version starts a server with a single endpoint: `/payment`.
* Added config parameters validation.
* Added `network_passphrase` config parameter.
* `postgres` migration files.
* Fixed sending to DiamNet address.
* Fixed `aurora.AccountResponse.SequenceNumber` bug.
* Fixed minor bugs.
* Code refactoring.
* Added example config file to the release package
* Updated README file.

## 0.0.1

* Initial release.
