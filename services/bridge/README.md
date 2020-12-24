# bridge-server
This is a stand alone server written in go. It is designed to make connecting to the DiamNet network as easy as possible. 
It allows you to be notified when a payment is received by a particular account. It also allows you to send a payment via a HTTP request.
It can optionally be connected to a `compliance` server if you want to carry out the compliance protocol.
It can be used by any project that needs to accept or send payments such as anchors or merchants accepting payments.

Handles:

- Creating DiamNet transactions.
- Monitoring a receiving DiamNet account.


## Downloading the server
[Prebuilt binaries](https://github.com/diamnet/bridge-server/releases) of the bridge-server server are available on the [releases page](https://github.com/diamnet/bridge-server/releases).

| Platform       | Binary file name                                                                         |
|----------------|------------------------------------------------------------------------------------------|
| Mac OSX 64 bit | [bridge-darwin-amd64](https://github.com/diamnet/bridge-server/releases)      |
| Linux 64 bit   | [bridge-linux-amd64](https://github.com/diamnet/bridge-server/releases)       |
| Windows 64 bit | [bridge-windows-amd64.exe](https://github.com/diamnet/bridge-server/releases) |

Alternatively, you can [build](#building) the binary yourself.

## Config

The `bridge.cfg` file must be present in a working directory (you can load another file by using `-c` parameter). Here is an [example configuration file](https://github.com/diamnet/bridge-server/blob/master/bridge_example.cfg). Config file should contain following values:

* `port` - server listening port
* `api_key` - when set, all requests to bridge server must contain `api_key` parameter with a correct value, otherwise the server will respond with `503 Forbidden`
* `network_passphrase` - passphrase of the network that will be used with this bridge server:
   * test network: `Test SDF Network ; September 2015`
   * public network: `Public Global DiamNet Network ; September 2015`
* `compliance` - URL to compliance server instance if you want to carry out the compliance protocol
* `aurora` - URL to [aurora](https://github.com/diamnet/aurora) server instance
* `assets` - array of approved assets codes that this server can authorize or receive. These are currency code/issuer pairs. Use asset code 'XLM' with no issuer to listen for XLM payments. See [`bridge_example.cfg`](./bridge_example.cfg) for example.
* `database`
  * `type` - database type (mysql, postgres)
  * `url` - url to database connection:
    * for `mysql`: `user:password@(host:port)/dbname` ([more info](https://github.com/go-sql-driver/mysql#dsn-data-source-name))
    * for `postgres`: `postgres://user:password@host/dbname?sslmode=sslmode` ([more info](https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters))
* `accounts`
  * `base_seed` - The secret seed of the account used to send payments. If left blank you will need to pass it in calls to `/payment`. 
  * `authorizing_seed` - The secret seed of the public key that is able to submit `allow_trust` operations on the issuing account.
  * `issuing_account_id` - The account ID of the issuing account (only if you want to authorize trustlines via bridge server, otherwise leave empty).
  * `receiving_account_id` - The account ID that receives incoming payments. The `callbacks.receive` will be called when a payment is received by this account.
* `callbacks`
  * `receive` - URL of the webhook where requests will be sent when a new payment is sent to the receiving account. The bridge server will keep calling the receive callback indefinitely until 200 OK status is returned by it. **WARNING** The bridge server can send multiple requests to this webhook for a single payment! You need to be prepared for it. See: [Security](#security).
  * `error` - URL of the webhook where requests will be sent when there is an error with an incoming payment
* `log_format` - set to `json` for JSON logs
* `mac_key` - a diamnet secret key used to add MAC headers to a payment notification.

Check [`bridge_example.cfg`](./bridge_example.cfg).

The minimal set of config values contains:
* `port`
* `network_passphrase`
* `aurora`

It will start a server with a single endpoint: `/payment`.

## Getting started

After creating `bridge.cfg` file, you need to run DB migrations:
```
./bridge --migrate-db
```

Then you can start the server:
```
./bridge
```

## API

`Content-Type` of requests data should be `application/x-www-form-urlencoded`.

### POST /create-keypair

Creates a new random key pair.

#### Response

```json
{
  "public_key": "GCSLLOYK7IKDQKUDSSAPHSJT3Y5XLIDIAFPVO5K42IN5CAQPNHIHJ2DE",
  "private_key": "SCJAOTWONWSOQLILCHNSGUOIXWCMIJQ563SPHMG25OPFX3IUDBAFU4SV"
}
```

In case of error it will return the following error:
* [`InternalServerError`](/src/github.com/diamnet/gateway/protocols/errors.go)

### POST /builder

Builds a transaction from a given request. `Content-Type` of this request should be `application/json`. Check [List of operations](https://www.diamnet.org/developers/learn/concepts/list-of-operations.html) doc to learn more about how each operation looks like.

**Note** This will not submit a transaction to the network. Please use [Aurora](https://www.diamnet.org/developers/aurora/reference/endpoints/transactions-create.html) to submit a transaction.

#### Request

Check example request below (remove comments before submitting it to the `bridge` server):

```json
{
  // Transaction source account
  "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
  // Sequence number
  "sequence_number": "123",
  // List of operations in this transaction
  "operations": [
    // First operation
    {
      // Operation type
      "type": "create_account",
      // Operation body
      "body": {
        // Don't send source field if operation source account is equal to transaction source account
        "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "destination": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "starting_balance": "50"
      }
    },
    // Second operation
    {
      "type": "payment",
      "body": {
        "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "destination": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "amount": "1050",
        "asset": {
          "code": "USD",
          "issuer": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT"
        }
      }
    },
    {
      "type": "path_payment",
      "body": {
        "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "destination": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "destination_amount": "10",
        "send_max": "1050",
        "send_asset": {
          "code": "USD",
          "issuer": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT"
        },
        "destination_asset": {
          "code": "EUR",
          "issuer": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT"
        },
        "path": [
          {}, // Native asset
          {
              "code": "ZAR",
              "issuer": "GBNIVKJTD2SMAXB5ALPBZ7CHRYYLCO5XSH55H6TI3Z37P7SCRXQVESG2"
            }
        ]
      }
    },
    {
        "type": "manage_offer",
        "body": {
          "selling": {
            "code": "EUR",
            "issuer": "GDOJMKTDLGGLROSSM5BV5MXIAQ3JZHASQFUV55WBJ45AFOUXSVVFGPTJ"
          },
          "buying": {
            "code": "USD",
            "issuer": "GACETOPHMOLSZLG5IQ3D6KQDKCAAYUYTTQHIEY6IGZE4VOBDD2YY6YAO"
          },
          "amount": "123456",
          "price": "2.93850088",
          "offer_id": "100"
        }
    },
    {
        "type": "create_passive_offer",
        "body": {
          "selling": {
            "code": "EUR",
            "issuer": "GDOJMKTDLGGLROSSM5BV5MXIAQ3JZHASQFUV55WBJ45AFOUXSVVFGPTJ"
          },
          "buying": {
            "code": "USD",
            "issuer": "GACETOPHMOLSZLG5IQ3D6KQDKCAAYUYTTQHIEY6IGZE4VOBDD2YY6YAO"
          },
          "amount": "123456",
          "price": "2.93850088"
        }
    },
    {
        "type": "set_options",
        "body": {
          "inflation_dest": "GBMPZVOMJ67WQBTBCVURDKTGL4557272EGQMAJCXPSMLOE63XPLL6SVA",
          "set_flags": [1, 2],
          "clear_flags": [4],
          "master_weight": 100,
          "low_threshold": 1,
          "medium_threshold": 2,
          "high_threshold": 3,
          "home_domain": "diamnet.org",
          "signer": {
            "public_key": "GA6VMJJQM2QBPPIXK2UVTAOS4XSSSAKSCOGFQE55IMRBQR65GIVDTTQV",
            "weight": 5
          }
        }
    },
    {
      "type": "change_trust",
      "body": {
        "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "asset": {
          "code": "USD",
          "issuer": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT"
        }
      }
    },
    {
      "type": "allow_trust",
      "body": {
        "source": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "trustor": "GBDCOZD7CHY26KS6ABEZPIJAMS2G7GP3YSTJ6DIRIQ6YUU77ZAPI2LVT",
        "asset_code": "USD",
        "authorize": true
      }
    },
    {
        "type": "account_merge",
        "body": {
          "destination": "GBLH67TQHRNRLERQEIQJDNBV2DSWPHAPP43MBIF7DVKA7X55APUNS4LL"
        }
    },
    {
        "type": "inflation",
        "body": {}
    },
    {
        "type": "manage_data",
        "body": {
          "name": "test_data",
          "data": "AQIDBAUG"
        }
    }
  ],
  // Array of signers
  "signers": ["SDOTALIMPAM2IV65IOZA7KZL7XWZI5BODFXTRVLIHLQZQCKK57PH5F3H"]
}
```

Assets are represented by a JSON object with two fields: `code` and `issuer`. Empty JSON object represents [native asset](https://www.diamnet.org/developers/learn/concepts/assets.html#lumens-xlm-).

#### Response

When transaction can be successfully built it will return a JSON object with a single `transaction_envelope` field that will contain base64-encoded `TransactionEnvelope` XDR object:

```json
{
    "transaction_envelope": "AAAAAEYnZH8R8a8qXgBJl6EgZLRvmfvEpp8NEUQ9i..."
}
```

In case of error it will return one of the following errors:
* [`InternalServerError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`InvalidParameterError`](/src/github.com/diamnet/gateway/protocols/errors.go)

### POST /payment

Builds and submits a transaction with a single [`payment`](https://www.diamnet.org/developers/learn/concepts/list-of-operations.html#payment), [`path_payment`](https://www.diamnet.org/developers/learn/concepts/list-of-operations.html#path-payment) or [`create_account`](https://www.diamnet.org/developers/learn/concepts/list-of-operations.html#create-account) (when sending native asset to account that does not exist) operation built from following parameters.

#### Safe transaction resubmittion

It’s possible that you will not receive a response from Bridge server due to a bug, network conditions, etc. In such situation it’s impossible to determine the status of your transaction and sending the same request to the Bridge server may result in "double-spend" of the funds. That’s why you should always send a request with `id` parameter set. Then when you resubmit a transaction with the same `id` the previously used [sequence number](https://www.diamnet.org/developers/guides/concepts/transactions.html) will be reused.

If the transaction has already been successfully applied to the ledger, Aurora server will simply return the saved result and not attempt to submit the transaction again. Only in cases where a transaction’s status is unknown (and thus will have a chance of being included into a ledger) will a resubmission to the network occur.

#### Request Parameters

Every request must contain required parameters from the following list. Additionally, depending on a type of payment, every request must contain required parameters for equivalent operation type.

name |  | description
--- | --- | ---
`id` | optional | Unique ID of the payment. If you send another request with the same `id` previously sent transaction will be resubmitted to the network. This parameter is required when sending a payment using Compliance protocol.
`source` | optional | Secret seed of transaction source account. If ommitted it will use the `base_seed` specified in the config file.
`sender` | optional | Payment address (ex. `bob*diamnet.org`) of payment sender account. Required for when sending using Compliance protocol.
`destination` | required | Account ID or payment address (ex. `bob*diamnet.org`) of payment destination account
`forward_destination[domain]` | required | Required when sending to Forward destination.
`forward_destination[fields][name]` | required | Required when sending to Forward destination. Fields will be added to Federation request query string.
`amount` | required | Amount that destination will receive
`memo_type` | optional | Memo type, one of: `id`, `text`, `hash`, `extra`
`memo` | optional | Memo value, `id` it must be uint64, when `hash` it must be 32 bytes hex value.
`use_compliance` | optional | When `true` Bridge will use Compliance protocol even if `extra_memo` is empty.
`extra_memo` | optional | You can include any info here and it will be included in the pre-image of the transaction's memo hash. See the [DiamNet Memo Convention](https://github.com/diamnet/diamnet-protocol/issues/28). When set and compliance server is connected, `memo` and `memo_type` values will be ignored.
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

##### Forward destination example

The following request to `/payment`:

```
forward_destination[domain]=diamnet.org&forward_destination[fields][forward_type]=bank_account&forward_destination[fields][swift]=BOPBPHMM&forward_destination[fields][acct]=2382376
```

will be translate to the following request:

```
https://FEDERATION_SERVER_READ_FROM_HCNET_TOML/federation?type=forward&forward_type=bank_account&swift=BOPBPHMM&acct=2382376
```

#### Response

It will return [`SubmitTransactionResponse`](/src/github.com/diamnet/gateway/aurora/submit_transaction_response.go) if there were no errors or with one of the following errors:

* [`InternalServerError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`InvalidParameterError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`MissingParameterError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`TransactionBadSequence`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionBadAuth`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionInsufficientBalance`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionNoAccount`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionInsufficientFee`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionBadAuthExtra`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`PaymentCannotResolveDestination`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentCannotUseMemo`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentSourceNotExist`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentAssetCodeNotAllowed`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentPending`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentDenied`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentMalformed`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentUnderfunded`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentSrcNoTrust`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentSrcNotAuthorized`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentNoDestination`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentNoTrust`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentNotAuthorized`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentLineFull`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentNoIssuer`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentTooFewOffers`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentOfferCrossSelf`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)
* [`PaymentOverSendmax`](/src/github.com/diamnet/gateway/protocols/bridge/payment.go)

#### Example

```sh
curl -X POST -d \
"source=SBNDIK4N7ZM3ZJKDJJDWDSPSRPHNI2RFL36WNNNEGQEW3G3AH6VJ2QB7&\
amount=1&\
destination=GBIUXI4S27PSL6TTJCJMPYDCF3K6AW2MYORFRTC7QBFE6NNEGVOQK46H&\
asset_code=USD&\
asset_issuer=GASZUHRFAFIZX5LR4WNHBWUXJBZNBEWCHFTR4XZHPF5TMVM5XUZBP5DT&\
memo_type=id&\
memo=125" \
http://localhost:8001/payment
```

### POST /authorize
Can be used to authorize other accounts to hold your assets.
It will build and submits a transaction with a [`allow_trust`](https://www.diamnet.org/developers/learn/concepts/list-of-operations.html#allow-trust) operation. 
The source of this transaction will be the account specified by `accounts.authorizing_seed` config parameter. 
You should make sure that this account is a low weight signer on the issuing account. See [Multi-sig](https://www.diamnet.org/developers/learn/concepts/multi-sig.html) for more information. 

#### Request Parameters

name |  | description
--- | --- | ---
`account_id` | required | Account ID of the account to authorize
`asset_code` | required | Asset code of the asset to authorize. Must be present in `assets` config array.

#### Response

It will return [`SubmitTransactionResponse`](/src/github.com/diamnet/gateway/aurora/submit_transaction_response.go) if there were no errors or with one of the following errors:

* [`InternalServerError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`InvalidParameterError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`MissingParameterError`](/src/github.com/diamnet/gateway/protocols/errors.go)
* [`TransactionBadSequence`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionBadAuth`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionInsufficientBalance`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionNoAccount`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionInsufficientFee`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`TransactionBadAuthExtra`](/src/github.com/diamnet/gateway/protocols/bridge/errors.go)
* [`AllowTrustMalformed`](/src/github.com/diamnet/gateway/protocols/bridge/authorize.go)
* [`AllowTrustNoTrustline`](/src/github.com/diamnet/gateway/protocols/bridge/authorize.go)
* [`AllowTrustTrustNotRequired`](/src/github.com/diamnet/gateway/protocols/bridge/authorize.go)
* [`AllowTrustCantRevoke`](/src/github.com/diamnet/gateway/protocols/bridge/authorize.go)

### POST /reprocess
Can be used to reprocess received payment.

#### Request Parameters

name |  | description
--- | --- | ---
`operation_id` | required | Aurora ID of operation to reprocess
`force` | optional | Must be set to `true` when reprocessing successful operations.

## Callbacks

The Bridge server listens for payment operations to the account specified by `accounts.receiving_account_id`. Every time 
a payment arrives it will send a HTTP POST request to `callbacks.receive`.

`Content-Type` of requests data will be `application/x-www-form-urlencoded`.

### `callbacks.receive`

The POST request with following parameters will be sent to this callback when a payment arrives.

> **Warning!** This callback can be called multiple times. Please check `id` parameter and respond with `200 OK` in case of duplicate payment.

#### Request

name | description
--- | ---
`id` | Operation ID (ex. `23110707918671873`)
`from` | Account ID of the sender
`route` | The recipient ID at the receiving FI. This will be the routing information contained in the memo or memo value if no compliance server is connected or memo type is not `hash`.
`amount` | Amount that was sent
`asset_code` | Code of the asset sent (ex. `USD`)
`asset_issuer` | Issuer of the asset sent (ex. `GD4I7AFSLZGTDL34TQLWJOM2NHLIIOEKD5RHHZUW54HERBLSIRKUOXRR`)
`memo_type` | Type of the memo attached to the transaction. This field will be empty when no memo was attached.
`memo` | Value of the memo attached. This field will be empty when no memo was attached.
`data` | Value of the [AuthData](https://www.diamnet.org/developers/learn/integration-guides/compliance-protocol.html). This field will be empty when compliance server is not connected.
`transaction_id` | The transaction hash of the operation (ex. `c7597583ad4f7caef15ad19b0f84017466b69790ee91bcacbbf98b51c93b17bf`)

#### Response

Respond with `200 OK` when processing succeeded. Any other status code will be considered an error and bridge server will keep sending this payment request again and will not continue to next payments until it receives `200 OK` response.

#### Payload Authentication

When the `mac_key` configuration value is set, the bridge server will attach HTTP headers to each payment notification that allow the receiver to verify that the notification is not forged.  A header named `X-Payload-Mac` that contains a base64-encoded MAC value will be included. This MAC is derived by calculating the HMAC-SHA256 of the raw request body using the decoded value of the `mac_key` configuration option as the key.

This MAC can be used on the receiving side of the notification to verify that the payment notifications was generated from the bridge server, rather than from some other actor, to increase security.

## Security

* This server must be set up in an isolated environment (ex. AWS VPC). Please make sure your firewall is properly configured 
and accepts connections from a trusted IPs only. You can set the `api_key` config parameter as an additional protection but it's not recommended as the solely protection. 
If you don't set this properly, an unauthorized person will be able to submit transactions from your accounts!
* Make sure the `callbacks` you provide only accept connections from the bridge server IP.
* Remember that `callbacks.receive` may be called multiple times with the same payment. Check `id` parameter and ignore 
requests with the same value (just send `200 OK` response).

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
http://localhost:6060/pkg/github.com/diamnet/gateway/
```
in a browser.
