# compliance-server
This is a stand alone server written in go. It is designed to make [Compliance protocol](https://www.hcnet.org/developers/learn/integration-guides/compliance-protocol.html) requests to other organizations. You can connect to it from the `bridge` server or any other server that can talk to it (check API section).

## Downloading the server
[Prebuilt binaries](https://github.com/hcnet/bridge-server/releases) of the bridge-server server are available on the [releases page](https://github.com/hcnet/bridge-server/releases).

| Platform       | Binary file name                                                                         |
|----------------|------------------------------------------------------------------------------------------|
| Mac OSX 64 bit | [bridge-darwin-amd64](https://github.com/hcnet/bridge-server/releases)      |
| Linux 64 bit   | [bridge-linux-amd64](https://github.com/hcnet/bridge-server/releases)       |
| Windows 64 bit | [bridge-windows-amd64.exe](https://github.com/hcnet/bridge-server/releases) |

Alternatively, you can [build](#building) the binary yourself.

## Config

The `compliance.cfg` file must be present in a working directory (you can load another file by using `-c` parameter). Here is an [example configuration file](https://github.com/hcnet/bridge-server/blob/master/compliance_example.cfg). Config file should contain following values:

* `external_port` - external server listening port (should be accessible from public)
* `internal_port` - internal server listening port (should be accessible from your internal network only!)
* `needs_auth` - set to `true` if you need to do sanctions check for payment receiver
* `network_passphrase` - passphrase of the network that will be used with this bridge server:
   * test network: `Test SDF Network ; September 2015`
   * public network: `Public Global HcNet Network ; September 2015`
* `database` - This database is used internally to store memo information and to keep track of what FIs have been authorized to receive customer info.
  * `type` - database type (mysql, postgres)
  * `url` - url to database connection. **IMPORTANT** The `compliance` server must not use the same database as the `bridge` server.
    * for `mysql`: `user:password@(host:port)/dbname` ([more info](https://github.com/go-sql-driver/mysql#dsn-data-source-name))
    * for `postgres`: `postgres://user:password@host/dbname?sslmode=sslmode` ([more info](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters))
* `keys`
  * `signing_seed` - The secret seed that will be used to sign messages. Public key derived from this secret key should be in your `hcnet.toml` file.
  * `encryption_key` - The secret key used to decrypt messages. _Not working yet._
* `callbacks`
  * `sanctions` - Callback that performs sanctions check. Read [Callbacks](#callbacks) section.
  * `ask_user` - Callback that asks user for permission for reading their data. Read [Callbacks](#callbacks) section.
  * `fetch_info` - Callback that returns user data. Read [Callbacks](#callbacks) section.
  * `tx_status` - Callback that returns user data. Read [Callbacks](#callbacks) section.
* `tls` (only when running HTTPS external server)
  * `certificate_file` - a file containing a certificate
  * `private_key_file` - a file containing a matching private key
* `log_format` - set to `json` for JSON logs
* `tx_status_auth` - authentication credentials for `/tx_status` endpoint.
  * `username`
  * `password` - minimum 10 chars

Check [`compliance_example.cfg`](./compliance_example.cfg).

## Getting started

After creating `compliance.cfg` file, you need to run DB migrations:
```
./compliance --migrate-db
```

Then you can start the server:
```
./compliance
```

## API

`Content-Type` of requests data should be `application/x-www-form-urlencoded`.

### POST :external_port/ (Auth endpoint)

Process auth request from external organization sent before sending a payment. Check [Compliance protocol](https://www.hcnet.org/developers/learn/integration-guides/compliance-protocol.html) for more info. It also saves memo preimage to the database.

#### Request Parameters

name |  | description
--- | --- | ---
`data` | required | Auth data.
`sig` | required | Signature.

Read more in [Compliance protocol](https://www.hcnet.org/developers/learn/integration-guides/compliance-protocol.html#auth_server) doc.

#### Response

Returns [Auth response](https://www.hcnet.org/developers/learn/integration-guides/compliance-protocol.html#reply).

### POST :internal_port/send

Typically called by the bridge server when a user initiates a payment. This endpoint causes the compliance server to send an Auth request to another organization. It will call the Auth endpoint of the receiving instition.

#### Request Parameters

name |  | description
--- | --- | ---
`id` | required | ID of the payment/transaction. In case of `pending` response or errors, you should resubmit the request with the same `id` value.
`source` | required | Account ID of transaction source account.
`sender` | required | HcNet address (ex. `bob*hcnet.org`) of payment sender account.
`destination` | required | Account ID or HcNet address (ex. `bob*hcnet.org`) of payment destination account
`amount` | required | Amount that destination will receive
`extra_memo` | optional | Additional information attached to memo preimage.
`asset_code` | optional | Asset code (XLM when empty) destination will receive
`asset_issuer` | optional | Account ID of asset issuer (XLM when empty) destination will receive
`send_max` | optional | [path_payment] Maximum amount of send_asset to send
`send_asset_code` | optional | [path_payment] Sending asset code (XLM when empty)
`send_asset_issuer` | optional | [path_payment] Account ID of sending asset issuer (XLM when empty)
`path[n][asset_code]` | optional | [path_payment] If the path isn't specified the bridge server will find the path for you. Asset code of `n`th asset on the path (XLM when empty, but empty parameter must be sent!)
`path[n][asset_issuer]` | optional | [path_payment] Account ID of `n`th asset issuer (XLM when empty, but empty parameter must be sent!)
`path[n+1][asset_code]` | optional | [path_payment] Asset code of `n+1`th asset on the path (XLM when empty, but empty parameter must be sent!)
`path[n+1][asset_issuer]` | optional | [path_payment] Account ID of `n+1`th asset issuer (XLM when empty, but empty parameter must be sent!)
... | ... | _Up to 5 assets in the path..._

#### Response

Returns [`SendResponse`]().

### POST :internal_port/receive

Typically called by the bridge server when a payment comes in. It is used to check that the payment was authorized by this compliance server. The call will return a memo preimage in the payment was authorized.

#### Request Parameters

name |  | description
--- | --- | ---
`memo` | required | Memo hash.

#### Response

Returns [`ReceiveResponse`]().

### POST :internal_port/allow_access

Allows access to users data for external user or FI.

#### Request Parameters

name |  | description
--- | --- | ---
`name` | required | Name of the external FI.
`domain` | required | Domain of the external FI.
`public_key` | required | Public key of the external FI.
`user_id` | optional | If set, only this user will be allowed.

#### Response

Will response with `200 OK` if saved. Any other status is an error.

### POST :internal_port/remove_access

Allows access to users data for external user or FI.

#### Request Parameters

name |  | description
--- | --- | ---
`domain` | required | Domain of the external FI.
`user_id` | optional | If set, only this user entry will be removed.

#### Response

Will response with `200 OK` if removed. Any other status is an error.

## Callbacks

The Compliance server will send callback `POST` request to URLs you define in the config file. `Content-Type` of requests data will be `application/x-www-form-urlencoded`.

### `callbacks.sanctions`

If set in the config file, this callback will be called when sanctions checks need to be performed. If not set the compliance server will act as if the sanction check passes.

#### Request

name | description
--- | ---
`sender` | Sender info JSON

The customer information that is exchanged between FIs is flexible but the typical fields are:

* Full Name
* Date of birth
* Physical address

#### Response

Respond with one of the following status codes:
* `200 OK` when sender/receiver is allowed and the payment should be processed,
* `202 Accepted` when your callback needs some time for processing,
* `400 Bad Request` when sender info is invalid.
* `403 Forbidden` when sender/receiver is denied.

Any other status code will be considered an error.

When `202 Accepted` is returned the response body should contain JSON object with `pending` field which represents the estimated number of seconds needed for processing. For example, the following response means to try the payment again in an hour.

When `400 Bad Request` is returned the response body should contain JSON object with `error` field with error string.

```json
{"pending": 3600}
```

### `callbacks.ask_user`

If set in the config file, this callback will be called when the sender needs your customer KYC info to send a payment. If not set then the customer information won't be given to the other FI.

#### Request

name | description
--- | ---
`amount` | Payment amount
`asset_code` | Payment asset code
`asset_issuer` | Payment asset issuer
`sender` | Sender info JSON
`note` | Note attached to the payment

The customer information (`sender`) that is exchanged between FIs is flexible but the typical fields are:

* Full Name
* Date of birth
* Physical address

#### Response

Respond with one of the following status codes:
* `200 OK` when your customer has allowed sharing his/her compliance information with the requesting FI.
* `202 Accepted` when your callback needs some time for processing, ie to ask the customer.
* `400 Bad Request` when request data is invalid.
* `403 Forbidden` when your customer has denied sharing his/her compliance information with the requesting FI.

Any other status code will be considered an error.

When `202 Accepted` is returned the response body should contain JSON object with `pending` field which represents estimated number of seconds needed for processing. For example, the following response means to try the payment again in an hour.

When `400 Bad Request` is returned the response body should contain JSON object with `error` field with error string.

```json
{"pending": 3600}
```

### `callbacks.fetch_info`

This callback should return the compliance information of your customer identified by `address`.

#### Request

name | description
--- | ---
`address` | HcNet address (ex. `alice*acme.com`) of the user.

#### Response

This callback should return `200 OK` status code and JSON object with the customer compliance info:

```json
{
  "name": "John Doe",
  "address": "User physical address",
  "date_of_birth": "1990-01-01"
}
```
### `callbacks.tx_status`
This callback should return the status of a transaction as explained in [`SEP-0004`](https://github.com/hcnet/hcnet-protocol/blob/master/ecosystem/sep-0004.md).

#### Request

name | description
--- | ---
`id` | HcNet transaction ID.

#### Response
This callback should return `200 OK` status code and JSON object with the transaction status info:

```json
{
  "status": "status code as defined in SEP-0001",
  "recv_code": "arbitrary string",
  "refund_tx": "tx_hash",
  "msg": "arbitrary string"
}
```


Any other status code will be considered an error.

## Building

[gb](http://getgb.io) is used for building and testing.

Given you have a running golang installation, you can build the server with:

```
gb build
```

After a successful build, you should find `bin/bridge` in the project directory.

## Running tests

```
gb test
```

## Documentation

```
godoc -goroot=. -http=:6060
```

Then simply open:
```
http://localhost:6060/pkg/github.com/hcnet/gateway/
```
in a browser.
