# Changelog

All notable changes to this project will be documented in this
file.  This project adheres to [Semantic Versioning](http://semver.org/).

As this project is pre 1.0, breaking changes may happen for minor version
bumps.  A breaking change will get clearly notified in this log.

## v0.20.1

* Add `--ingest-state-reader-temp-set` flag (`INGEST_STATE_READER_TEMP_SET` env variable) which defines the storage type used for temporary objects during state ingestion in the new ingestion system. The possible options are: `memory` (requires ~1.5GB RAM, fast) and `postgres` (stores data in temporary table in Postgres, less RAM but slower).

## v0.20.0

### Changes

* Experimental ingestion system is now run concurrently on all Aurora servers (with feature flag set - see below). This improves ingestion availability.
* Add experimental path finding endpoints which use an in memory order book for improved performance. To enable the endpoints set `--enable-experimental-ingestion` CLI param or `ENABLE_EXPERIMENTAL_INGESTION=true` env variable. Note that the `enable-experimental-ingestion` flag enables both the new path finding endpoints and the accounts for signer endpoint. There are two path finding endpoints. `/paths/strict-send` returns payment paths where both the source and destination assets are fixed. This endpoint is able to answer questions like: "Get me the most EUR possible for my 10 USD." `/paths/strict-receive` is the other endpoint which is an alias to the existing `/paths` endpoint. 
* `--enable-accounts-for-signer` CLI param or `ENABLE_ACCOUNTS_FOR_SIGNER=true` env variable are merged with `--enable-experimental-ingestion` CLI param or `ENABLE_EXPERIMENTAL_INGESTION=true` env variable.
* Add experimental get offers by id endpoint`/offers/{id}` which uses the new ingestion system to fill up the offers table. To enable it, set `--enable-experimental-ingestion` CLI param or `ENABLE_EXPERIMENTAL_INGESTION=true` env variable.

### Scheduled Breaking Changes

* `fee_paid` field on Transaction resource has been deprecated and will be removed in 0.22.0. Please use new fields added in 0.18.0: `max_fee` that defines the maximum fee the source account is willing to pay and `fee_charged` that defines the fee that was actually paid for a transaction. See [CAP-0005](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0005.md) for more information. This change has been previously scheduled for 0.19.0 release.
* The following operation type names have been deprecated: `manage_offer` and `create_passive_offer`. The names will be changed to: `manage_sell_offer` and `create_passive_offer` in 0.22.0. This has been previously scheduled for 0.19.0 release.
* The type for the following attributes will be changed from `int64` to `string` in 0.22.0:
  - Attribute `offer_id` in [manage buy offer](https://www.diamnet.org/developers/aurora/reference/resources/operation.html#manage-buy-offer) and [manage sell offer](https://www.diamnet.org/developers/aurora/reference/resources/operation.html#manage-sell-offer) operations.
  - Attribute `offer_id` in `Trade` effect.
  - Attribute `id` in [Offer](https://www.diamnet.org/developers/aurora/reference/resources/offer.html) resource.
  - Attribute `timestamp` and `trade_count` in [Trade Aggregation](https://www.diamnet.org/developers/aurora/reference/resources/trade_aggregation.html) resource.

If you are an SDK maintainer, update your code to prepare for this change.

## v0.19.0

* Add `join` parameter to operations and payments endpoints. Currently, the only valid value for the parameter is `transactions`. If `join=transactions` is included in a request then the response will include a `transaction` field for each operation in the response.
* Add experimental "Accounts For Signers" endpoint. To enable it set `--enable-accounts-for-signer` CLI param or `ENABLE_ACCOUNTS_FOR_SIGNER=true` env variable. Additionally new feature requires links to history archive: CLI: `--history-archive-urls="archive1,archive2,archive3"`, env variable: `HISTORY_ARCHIVE_URLS="archive1,archive2,archive3"`. This will expose `/accounts` endpoint. This requires around 4GB of RAM for initial state ingestion.

## v0.18.1

* Fixed `/fee_stats` to correctly calculate ledger capacity in protocol v11.
* Fixed `aurora db clean` command to truncate all tables.

## v0.18.0

### Breaking changes

* Aurora requires Postgres 9.5+.
* Removed `paging_token` field from `/accounts/{id}` endpoint.
* Removed `/operation_fee_stats` endpoint. Please use `/fee_stats`.

### Deprecations

* `fee_paid` field on Transaction resource has been deprecated and will be removed in 0.19.0. Two new fields have been added: `max_fee` that defines the maximum fee the source account is willing to pay and `fee_charged` that defines the fee that was actually paid for a transaction. See [CAP-0005](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0005.md) for more information.
* The following operation type names have been deprecated: `manage_offer` and `create_passive_offer`. The names will be changed to: `manage_sell_offer` and `create_passive_offer` in 0.19.0.

### Changes

* The following new config parameters were added. When old `max-db-connections` config parameter is set, it has a priority over the the new params. Run `aurora help` for more information.
  * `aurora-db-max-open-connections`,
  * `aurora-db-max-idle-connections`,
  * `core-db-max-open-connections`,
  * `core-db-max-idle-connections`.
* Fixed `fee_paid` value in Transaction resource (#1358).
* Fix "int64: value out of range" errors in trade aggregations (#1319).
* Improved `aurora db reingest range` command.

## v0.17.6 - 2019-04-29

* Fixed a bug in `/order_book` when sum of amounts at a single price level exceeds `int64_max` (#1037).
* Fixed a bug generating `ERROR` level log entries for bad requests (#1186).

## v0.17.5 - 2019-04-24

* Support for diamnet-core [v11.0.0](https://github.com/diamnet/diamnet-core/releases/tag/v11.0.0).
* Display trustline authorization state in the balances list.
* Improved actions code.
* Improved `aurora db reingest` command handling code.
* Tracking app name and version that connects to Aurora (`X-App-Name`, `X-App-Version`).

## v0.17.4 - 2019-03-14

* Support for DiamNet-Core 10.3.0 (new database schema v9).
* Fix a bug in `aurora db reingest` command (no log output).
* Multiple code improvements.

## v0.17.3 - 2019-03-01

* Fix a bug in `txsub` package that caused returning invalid status when resubmitting old transactions (#969).

## v0.17.2 - 2019-02-28

* Critical fix bug

## v0.17.1 - 2019-02-28

### Changes

* Fixes high severity error in ingestion system.
* Account detail endpoint (`/accounts/{id}`) includes `last_modified_ledger` field for account and for each non-native asset balance.

## v0.17.0 - 2019-02-26

### Upgrade notes

This release introduces ingestion of failed transactions. This feature is turned off by default. To turn it on set environment variable: `INGEST_FAILED_TRANSACTIONS=true` or CLI param: `--ingest-failed-transactions=true`. Please note that ingesting failed transactions can double DB space requirements (especially important for full history deployments).

### Database migration notes

Previous versions work fine with new schema so you can migrate (`aurora db migrate up` using new binary) database without stopping the Aurora process. To reingest ledgers run `aurora db reingest` using Aurora 0.17.0 binary. You can take advantage of the new `aurora db reingest range` for parallel reingestion.

### Deprecations

* `/operation_fee_stats` is deprecated in favour of `/fee_stats`. Will be removed in v0.18.0.

### Breaking changes

* Fields removed in this version:
  * Root > `protocol_version`, use `current_protocol_version` and `core_supported_protocol_version`.
  * Ledger > `transaction_count`, use `successful_transaction_count` and `failed_transaction_count`.
  * Signer > `public_key`, use `key`.
* This Aurora version no longer supports Core <10.0.0. Aurora can still ingest version <10 ledgers.
* Error event name during streaming changed to `error` to follow W3C specification.

### Changes

* Added ingestion of failed transactions (see Upgrade notes). Use `include_failed=true` GET parameter to display failed transactions and operations in collection endpoints.
* `/fee_stats` endpoint has been extended with fee percentiles and ledger capacity usage. Both are useful in transaction fee estimations.
* Fixed a bug causing slice bounds out of range at `/account/{id}/offers` endpoint during streaming.
* Added `aurora db reingest range X Y` that reingests ledgers between X and Y sequence number (closed intervals).
* Many code improvements.

## v0.16.0 - 2019-02-04

### Upgrade notes

* Ledger > Admins need to reingest old ledgers because we introduced `successful_transaction_count` and `failed_transaction_count`.

### Database migration notes

Previous versions work fine with Aurora 0.16.0 schema so you can migrate (`aurora db migrate up`) database without stopping the Aurora process. To reingest ledgers run `aurora db reingest` using Aurora 0.16.0 binary.

### Deprecations

* Root > `protocol_version` will be removed in v0.17.0. It is replaced by `current_protocol_version` and `core_supported_protocol_version`.
* Ledger > `transaction_count` will be removed in v0.17.0.
* Signer > `public_key` will be removed in v0.17.0.

### Changes

* Improved `aurora db migrate` script. It will now either success or show a detailed message regarding why it failed.
* Fixed effects ingestion of circular payments.
* Improved account query performances for payments and operations.
* Added `successful_transaction_count` and `failed_transaction_count` to `ledger` resource.
* Fixed the wrong protocol version displayed in `root` resource by adding `current_protocol_version` and `core_supported_protocol_version`.
* Improved streaming for single objects. It won't send an event back if the current event is the same as the last event sent.
* Fixed ingesting effects of empty trades. Empty trades will be ignored during ingestion.

## v0.15.4 - 2019-01-17

* Fixed multiple issues in transaction submission subsystem.
* Support for client fingerprint headers.
* Fixed parameter checking in `aurora db backfill` command.

## v0.15.3 - 2019-01-07

* Fixed a bug in Aurora DB reaping code.
* Fixed query checking code that generated `ERROR`-level log entries for invalid input.

## v0.15.2 - 2018-12-13

* Added `aurora db init-asset-stats` command to initialize `asset_stats` table. This command should be run once before starting ingestion if asset stats are enabled (`ENABLE_ASSET_STATS=true`).
* Fixed `asset_stats` table to support longer `home_domain`s.
* Fixed slow trades DB query.

## v0.15.1 - 2018-11-09

* Fixed memory leak in SSE stream code.

## v0.15.0 - 2018-11-06

DB migrations add a new fields and indexes on `history_trades` table. This is a very large table in `CATCHUP_COMPLETE` deployments so migration may take a long time (depending on your DB hardware). Please test the migrations execution time on the copy of your production DB first.

This release contains several bug fixes and improvements:

* New `/operation_fee_stats` endpoint includes fee stats for the last 5 ledgers.
* ["Trades"](https://www.diamnet.org/developers/aurora/reference/endpoints/trades.html) endpoint can now be streamed.
* In ["Trade Aggregations"](https://www.diamnet.org/developers/aurora/reference/endpoints/trade_aggregations.html) endpoint, `offset` parameter has been added.
* Path finding bugs have been fixed and the algorithm has been improved. Check [#719](https://github.com/diamnet/go/pull/719) for more information.
* Connections (including streams) are closed after timeout defined using `--connection-timeout` CLI param or `CONNECTION_TIMEOUT` environment variable. If Aurora is behind a load balancer with idle timeout set, it is recommended to set this to a value equal a few seconds less than idle timeout so streams can be properly closed by Aurora.
* Streams have been improved to check for updates every `--sse-update-frequency` CLI param or `SSE_UPDATE_FREQUENCY` environment variable seconds. If a new ledger has been closed in this period, new events will be sent to a stream. Previously streams checked for new events every 1 second, even when there were no new ledgers.
* Rate limiting algorithm has been changed to [GCRA](https://brandur.org/rate-limiting#gcra).
* Rate limiting in streams has been changed to be more fair. Now 1 *credit* has to be *paid* every time there's a new ledger instead of per request.
* Rate limiting can be disabled completely by setting `--per-hour-rate-limit=0` CLI param or `PER_HOUR_RATE_LIMIT=0` environment variable.
* Account flags now display `auth_immutable` value.
* Logs can be sent to a file. Destination file can be set using an environment variable (`LOG_FILE={file}`) or CLI parameter (`--log-file={file}`).

### Breaking changes

* Assets stats are disabled by default. This can be changed using an environment variable (`ENABLE_ASSET_STATS=true`) or CLI parameter (`--enable-asset-stats=true`). Please note that it has a negative impact on a DB and ingestion time.
* In ["Offers for Account"](https://www.diamnet.org/developers/aurora/reference/endpoints/offers-for-account.html), `last_modified_time` field  endpoint can be `null` when ledger data is not available (has not been ingested yet).
* ["Trades for Offer"](https://www.diamnet.org/developers/aurora/reference/endpoints/trades-for-offer.html) endpoint will query for trades that match the given offer on either side of trades, rather than just the "sell" offer. Offer IDs are now [synthetic](https://www.diamnet.org/developers/aurora/reference/resources/trade.html#synthetic-offer-ids). You have to reingest history to update offer IDs.

### Other bug fixes

* `aurora db backfill` command has been fixed.
* Fixed `remoteAddrIP` function to support IPv6.
* Fixed `route` field in the logs when the request is rate limited.

## v0.14.2 - 2018-09-27

### Bug fixes

* Fixed and improved `txsub` package (#695). This should resolve many issues connected to `Timeout` responses.
* Improve stream error reporting (#680).
* Checking `ingest.Cursor` errors in `Session` (#679).
* Added account ID validation in `/account/{id}` endpoints (#684).

## v0.14.1 - 2018-09-19

This release contains several bug fixes:

* Assets stats can cause high CPU usage on diamnet-core DB. If this slows down the database it's now possible to turn off this feature by setting `DISABLE_ASSET_STATS` feature flag. This can be set as environment variable (`DISABLE_ASSET_STATS=true`) or CLI parameter (`--disable-asset-stats=true`).
* Sometimes `/accounts/{id}/offers` returns `500 Internal Server Error` response when ledger data is not available yet (for new ledgers) or no longer available (`CATCHUP_RECENT` deployments). It's possible to set `ALLOW_EMPTY_LEDGER_DATA_RESPONSES` feature flag as environment variable (`ALLOW_EMPTY_LEDGER_DATA_RESPONSES=true`) or CLI parameter (`--allow-empty-ledger-data-responses=true`). With the flag set to `true` "Offers for Account" endpoint will return `null` in `last_modified_time` field when ledger data is not available, instead of `500 Internal Server Error` error.

### Bug fixes

* Feature flag to disable asset stats (#668).
* Feature flag to allow null ledger data in responses (#672).
* Fix empty memo field in JSON when memo_type is text (#635).
* Improved logging: some bad requests no longer generate `ERROR` level log entries (#654).
* `/friendbot` endpoint is available only when `FriendbotURL` is set in the config.

## v0.14.0 - 2018-09-06

### Breaking changes

* Offer resource `last_modified` field removed (see Bug Fixes section).
* Trade aggregations endpoint accepts only specific time ranges now (1/5/15 minutes, 1 hour, 1 day, 1 week).
* Aurora sends `Cache-Control: no-cache, no-store, max-age=0` HTTP header for all responses.

### Deprecations

* Account > Signers collection `public_key` field is deprecated, replaced by `key`.

### Changes

* Protocol V10 features:
  * New `bump_sequence` operation (as in [CAP-0001](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0001.md)).
    * New [`bump_sequence`](https://www.diamnet.org/developers/aurora/reference/resources/operation.html#bump-sequence) operation.
    * New `sequence_bumped` effect.
    * Please check [CAP-0001](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0001.md) for new error codes for transaction submission.
  * Offer liabilities (as in [CAP-0003](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0003.md)):
    * `/accounts/{id}` resources contain new fields: `buying_liabilities` and `selling_liabilities` for each entry in `balances`.
    * Please check [CAP-0003](https://github.com/diamnet/diamnet-protocol/blob/master/core/cap-0003.md) for new error codes for transaction submission.
* Added `source_amount` field to `path_payment` operations.
* Added `account_credited` and `account_debited` effects for `path_payment` operations.
* Friendbot link in Root endpoint is empty if not set in configuration.
* Improved `ingest` package logging.
* Improved HTTP logging (`forwarded_ip`, `route` fields, `duration` is always in seconds).
* `LOGGLY_HOST` env variable has been replaced with `LOGGLY_TAG` and is adding a tag to every log event.
* Dropped support for Go 1.8.

### Bug fixes

* New fields in Offer resource: `last_modified_ledger` and `last_modified_time`, replace buggy `last_modified` (#478).
* Fixed pagination in Trades for account endpoint (#486).
* Fixed a synchronization issue in `ingest` package (#603).
* Fixed Order Book resource links in Root endpoint.
* Fixed streaming in Offers for Account endpoint.

## v0.13.3 - 2018-08-23

### Bug fixes

* Fixed large amounts rendering in `/assets`.

## v0.13.2 - 2018-08-13

### Bug fixes

* Fixed a bug in `amount` and `price` packages triggering long calculations.

## v0.13.1 - 2018-07-26

### Bug fixes

* Fixed a conversion bug when `timebounds.max_time` is set to `INT64_MAX`.

## v0.13.0 - 2018-06-06

### Breaking changes

- `amount` field in `/assets` is now a String (to support DiamNet amounts larger than `int64`).

### Changes

- Effect resource contains a new `created_at` field.
- Aurora responses are compressed.
- Ingestion errors have been improved.
- `aurora rebase` command was improved.

### Bug fixes

- Aurora now returns `400 Bad Request` for negative `cursor` values.

**Upgrade notes**

DB migrations add a new indexes on `history_trades`. This is very large table so migration may take a long time (depending on your DB hardware). Please test the migrations execution time on the copy of your production DB first.

## v0.12.3 - 2017-03-20

### Bug fixes

- Fix a service stutter caused by excessive `info` commands being issued from the root endpoint.


## v0.12.2 - 2017-03-14

This release is a bug fix release for v0.12.1 and v0.12.2.  *Please see the upgrade notes below if you did not already migrate your db for v0.12.0*

### Changes

- Remove strict validation on the `resolution` parameter for trade aggregations endpoint.  We will add this feature back in to the next major release. 


## v0.12.1 - 2017-03-13

This release is a bug fix release for v0.12.0.  *Please see the upgrade notes below if you did not already migrate your db for v0.12.0*

### Bug fixes

- Fixed an issue caused by un-migrated trade rows. (https://github.com/diamnet/go/issues/357)
- Command line flags are now useable for subcommands of aurora.


## v0.12.0 - 2017-03-08

Big release this time for aurora:  We've made a number of breaking changes since v0.11.0 and have revised both our database schema as well as our data ingestion system.  We recommend that you take a backup of your aurora database prior to upgrading, just in case.  

### Upgrade Notes

Since this release changes both the schema and the data ingestion system, we recommend the following upgrade path to minimize downtime:

1. Upgrade aurora binaries, but do not restart the service
2. Run `aurora db migrate up` to migrate the db schema
3. Run `aurora db reingest` in a background session to begin the data reingestion process
4. Restart aurora

### Added

- Operation and payment resources were changed to add `transaction_hash` and `created_at` properties.
- The ledger resource was changed to add a `header_xdr` property.  Existing aurora installations should re-ingest all ledgers to populate the history database tables with the data.  In future versions of aurora we will disallow null values in this column.  Going forward, this change reduces the coupling of aurora to diamnet-core, ensuring that aurora can re-import history even when the data is no longer stored within diamnet-core's database.
- All Assets endpoint (`/assets`) that returns a list of all the assets in the system along with some stats per asset. The filters allow you to narrow down to any specific asset of interest.
- Trade Aggregations endpoint (`/trade_aggregations`) allow for efficient gathering of historical trade data. This is done by dividing a given time range into segments and aggregate statistics, for a given asset pair (`base`, `counter`) over each of these segments.

### Bug fixes

- Ingestion performance and stability has been improved. 
- Changes to an account's inflation destination no longer produce erroneous "signer_updated" effects. (https://github.com/diamnet/aurora/issues/390)


### Changed

- BREAKING CHANGE: The `base_fee` property of the ledger resource has been renamed to `base_fee_in_stroops` 
- BREAKING CHANGE: The `base_reserve` property of the ledger resource has been renamed to `base_reserve_in_stroops` and is now expressed in stroops (rather than lumens) and as a JSON number. 
- BREAKING CHANGE: The "Orderbook Trades" (`/orderbook/trades`) endpoint has been removed and replaced by the "All Trades" (`/trades`) endpoint.
- BREAKING CHANGE: The Trade resource has been modified to generalize assets as (`base`, `counter`) pairs, rather than the previous (`sold`,`bought`) pairs.  
- Full reingestion (i.e. running `aurora db reingest`) now runs in reverse chronological order.  

### Removed

- BREAKING CHANGE: Friendbot has been extracted to an external microservice.


## [v0.11.0] - 2017-08-15

### Bug fixes

- The ingestion system can now properly import envelopes that contain signatures that are zero-length strings.
- BREAKING CHANGE: specifying a `limit` of `0` now triggers an error instead of interpreting the value to mean "use the default limit".
- Requests that ask for more records than the maximum page size now trigger a bad request error, instead of an internal server error.
- Upstream bug fixes to xdr decoding from `github.com/diamnet/go`.

### Changed

- BREAKING CHANGE: The payments endpoint now includes `account_merge` operations in the response.
- "Finished Request" log lines now include additional fields: `streaming`, `path`, `ip`, and `host`.
- Responses now include a `Content-Disposition: inline` header.


## [v0.10.1] - 2017-03-29

### Fixed
- Ingestion was fixed to protect against text memos that contain null bytes.  While memos with null bytes are legal in diamnet-core, PostgreSQL does not support such values in string columns.  Aurora now strips those null bytes to fix the issue. 

## [v0.10.0] - 2017-03-20

This is a fix release for v0.9.0 and v0.9.1


### Added
- Added `aurora db clear` helper command to clear previously ingested history.

### Fixed

- Embedded sql files for the database schema have been fixed agsain to be compatible with postgres 9.5. The configuration setting `row_security` has been removed from the dumped files.

## [v0.9.1] - 2017-03-20

### Fixed

- Embedded sql files for the database schema have been fixed to be compatible with postgres 9.5. The configuration setting `idle_in_transaction_session_timeout` has been removed from the dumped files.

## [v0.9.0] - 2017-03-20

This release was retracted due to a bug discovered after release.

### Added
- Aurora now exposes the diamnet network protocol in several places:  It shows the currently reported protocol version (as returned by the diamnet-core `info` command) on the root endpoint, and it reports the protocol version of each ledger resource.
- Trade resources now include a `created_at` timestamp.

### Fixed

- BREAKING CHANGE: The reingestion process has been updated.  Prior versions of aurora would enter a failed state when a gap between the imported history and the diamnet-core database formed or when a previously imported ledger was no longer found in the diamnet-core database.  This usually occurs when running diamnet-core with the `CATCHUP_RECENT` config option.  With these changed, aurora will automatically trim "abandonded" ledgers: ledgers that are older than the core elder ledger.


## [v0.8.0] - 2017-02-07

### Added

- account signer resources now contain a type specifying the type of the signer: `ed25519_public_key`, `sha256_hash`, and `preauth_tx` are the present values used for the respective signer types.

### Changed

- The `public_key` field on signer effects and an account's signer summary has been renamed to `key` to reflect that new signer types are not necessarily specifying a public key anymore.

### Deprecated

- The `public_key` field on account signers and signer effects are deprecated

## [v0.7.1] - 2017-01-12

### Bug fixes

- Trade resources now include "bought_amount" and "sold_amount" fields when being viewed through the "Orderbook Trades" endpoint.
- Fixes #322: orderbook summaries with over 20 bids now return the correct price levels, starting with the closest to the spread.

## [v0.7.0] - 2017-01-10

### Added

- The account resource now includes links to the account's trades and data values.

### Bug fixes

- Fixes paging_token attribute of account resource
- Fixes race conditions in friendbot
- Fixes #202: Add price and price_r to "manage_offer" operation resources
- Fixes #318: order books for the native currency now filters correctly.

## [v0.6.2] - 2016-08-18

### Bug fixes

- Fixes streaming (SSE) requests, which were broken in v0.6.0

## [v0.6.1] - 2016-07-26

### Bug fixes

- Fixed an issue where accounts were not being properly returned when the  history database had no record of the account.


## [v0.6.0] - 2016-07-20

This release contains the initial implementation of the "Abridged History System".  It allows a aurora system to be operated without complete knowledge of the ledger's history.  With this release, aurora will start ingesting data from the earliest point known to the connected diamnet-core instance, rather than ledger 1 as it behaved previously.  See the admin guide section titled "Ingesting diamnet-core data" for more details.

### Added

- *Elder* ledgers have been introduced:  An elder ledger is the oldest ledger known to a db.  For example, the `core_elder_ledger` attribute on the root endpoint refers to the oldest known ledger stored in the connected diamnet-core database.
- Added the `history-retention-count` command line flag, used to specify the amount of historical data to keep in the history db.  This is expressed as a number of ledgers, for example a value of `362880` would retain roughly 6 weeks of data given an average of 10 seconds per ledger.
- Added the `history-stale-threshold` command line flag to enable stale history protection.  See the admin guide for more info.
- Aurora now reports the last ledger ingested to diamnet-core using the `setcursor` command.
- Requests for data that precede the recorded window of history stored by aurora will receive a `410 Gone` http response to allow software to differentiate from other "not found" situations.
- The new `db reap` command will manually trigger the deletion of unretained historical data
- A background process on the server now deletes unretained historical once per hour.

### Changed

- BREAKING: When making a streaming request, a normal error response will be returned if an error occurs prior to sending the first event.  Additionally, the initial http response and streaming preamble will not be sent until the first event is available.
- BREAKING: `aurora_latest_ledger` has renamed to `history_latest_ledger`
- Aurora no longer needs to begin the ingestion of historical data from ledger sequence 1.  
- Rows in the `history_accounts` table are no longer identified using the "Total Order ID" that other historical records  use, but are rather using a simple auto-incremented id.

### Removed

- The `/accounts` endpoint, which lets a consumer page through the entire set of accounts in the ledger, has been removed.  The change from complete to an abridged history in aurora makes the endpoint mostly useless, and after consulting with the community we have decided to remove the endpoint.

## [v0.5.1] - 2016-04-28

### Added

  - ManageData operation data is now rendered in the various operation end points.

### Bug fixes

- Transaction memos that contain utf-8 are now properly rendered in browsers by properly setting the charset of the http response.

## [v0.5.0] - 2016-04-22

### Added

- BREAKING: Aurora can now import data from diamnet-core without the aid of the aurora-importer project.  This process is now known as "ingestion", and is enabled by either setting the `INGEST` environment variable to "true" or specifying "--ingest" on the launch arguments for the aurora process.  Only one process should be running in this mode for any given aurora database.
- Add `aurora db init`, used to install the latest bundled schema for the aurora database.
- Add `aurora db reingest` command, used to update outdated or corrupt aurora database information.  Admins may now use `aurora db reingest outdated` to migrate any old data when updated aurora.
- Added `network_passphrase` field to root resource.
- Added `fee_meta_xdr` field to transaction resource.

### Bug fixes
- Corrected casing on the "offers" link of an account resource.

## [v0.4.0] - 2016-02-19

### Added

- Add `aurora db migrate [up|down|redo]` commands, used for installing schema migrations.  This work is in service of porting the aurora-importer project directly to aurora.
- Add support for TLS: specify `--tls-cert` and `--tls-key` to enable.
- Add support for HTTP/2.  To enable, use TLS.

### Removed

- BREAKING CHANGE: Removed support for building on go versions lower than 1.6

## [v0.3.0] - 2016-01-29

### Changes

- Fixed incorrect `source_amount` attribute on pathfinding responses.
- BREAKING CHANGE: Sequence numbers are now encoded as strings in JSON responses.
- Fixed broken link in the successful response to a posted transaction

## [v0.2.0] - 2015-12-01
### Changes

- BREAKING CHANGE: the `address` field of a signer in the account resource has been renamed to `public_key`.
- BREAKING CHANGE: the `address` on the account resource has been renamed to `account_id`.

## [v0.1.1] - 2015-12-01

### Added
- Github releases are created from tagged travis builds automatically

[v0.11.0]: https://github.com/diamnet/aurora/compare/v0.10.1...v0.11.0
[v0.10.1]: https://github.com/diamnet/aurora/compare/v0.10.0...v0.10.1
[v0.10.0]: https://github.com/diamnet/aurora/compare/v0.9.1...v0.10.0
[v0.9.1]: https://github.com/diamnet/aurora/compare/v0.9.0...v0.9.1
[v0.9.0]: https://github.com/diamnet/aurora/compare/v0.8.0...v0.9.0
[v0.8.0]: https://github.com/diamnet/aurora/compare/v0.7.1...v0.8.0
[v0.7.1]: https://github.com/diamnet/aurora/compare/v0.7.0...v0.7.1
[v0.7.0]: https://github.com/diamnet/aurora/compare/v0.6.2...v0.7.0
[v0.6.2]: https://github.com/diamnet/aurora/compare/v0.6.1...v0.6.2
[v0.6.1]: https://github.com/diamnet/aurora/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/diamnet/aurora/compare/v0.5.1...v0.6.0
[v0.5.1]: https://github.com/diamnet/aurora/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/diamnet/aurora/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/diamnet/aurora/compare/v0.3.0...v0.4.0
[v0.3.0]: https://github.com/diamnet/aurora/compare/v0.2.0...v0.3.0
[v0.2.0]: https://github.com/diamnet/aurora/compare/v0.1.1...v0.2.0
[v0.1.1]: https://github.com/diamnet/aurora/compare/v0.1.0...v0.1.1
