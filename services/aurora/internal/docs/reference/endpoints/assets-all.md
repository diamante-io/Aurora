---
title: All Assets
clientData:
  laboratoryUrl:
---

This endpoint represents all [assets](../resources/asset.md).
It will give you all the assets in the system along with various statistics about each.

Note: When running this in `catchup_recent` mode you will only get a subset of all the assets in the system.
This is because we only register assets when they are encountered during ingestion.

## Request

```
GET /assets{?asset_code,asset_issuer,cursor,limit,order}
```

### Arguments

| name | notes | description | example |
| ---- | ----- | ----------- | ------- |
| `?asset_code` | optional, string, default _null_ | Code of the Asset to filter by | `USD` |
| `?asset_issuer` | optional, string, default _null_ | Issuer of the Asset to filter by | `GA2HGBJIJKI6O4XEM7CZWY5PS6GKSXL6D34ERAJYQSPYA6X6AI7HYW36` |
| `?cursor` | optional, any, default _null_ | A paging token, specifying where to start returning records from. | `1` |
| `?order` | optional, string, default `asc` | The order in which to return rows, "asc" or "desc", ordered by asset_code then by asset_issuer. | `asc` |
| `?limit` | optional, number, default: `10` | Maximum number of records to return. | `200` |

### curl Example Request

```sh
# Retrieve the 200 newest assets, ordered chronologically:
curl "https://aurora-testnet.diamnet.org/assets?limit=200&order=desc"
```

### JavaScript Example Request

```javascript
var DiamNetSdk = require('diamnet-sdk');
var server = new DiamNetSdk.Server('https://aurora-testnet.diamnet.org');

server.assets()
  .call()
  .then(function (result) {
    console.log(result.records);
  })
  .catch(function (err) {
    console.log(err)
  })
```

## Response

If called normally this endpoint responds with a [page](../resources/page.md) of assets.

### Example Response

```json
{
  "_links": {
    "self": {
      "href": "/assets?order=asc\u0026limit=10\u0026cursor="
    },
    "next": {
      "href": "/assets?order=asc\u0026limit=10\u0026cursor=3"
    },
    "prev": {
      "href": "/assets?order=desc\u0026limit=10\u0026cursor=1"
    }
  },
  "_embedded": {
    "records": [
      {
        "_links": {
          "toml": {
            "href": "https://www.diamnet.org/.well-known/diamnet.toml"
          }
        },
        "asset_type": "credit_alphanum12",
        "asset_code": "BANANA",
        "asset_issuer": "GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN",
        "paging_token": "BANANA_GDSBCQO34HWPGUGQSP3QBFEXVTSR2PW46UIGTHVWGWJGQKH3AFNHXHXN_credit_alphanum4",
        "amount": "10000.0000000",
        "num_accounts": 2126,
        "flags": {
          "auth_required": true,
          "auth_revocable": false
        }
      },
      {
        "_links": {
          "toml": {
            "href": "https://www.diamnet.org/.well-known/diamnet.toml"
          }
        },
        "asset_type": "credit_alphanum4",
        "asset_code": "BTC",
        "asset_issuer": "GBAUUA74H4XOQYRSOW2RZUA4QL5PB37U3JS5NE3RTB2ELJVMIF5RLMAG",
        "paging_token": "BTC_GBAUUA74H4XOQYRSOW2RZUA4QL5PB37U3JS5NE3RTB2ELJVMIF5RLMAG_credit_alphanum4",
        "amount": "5000.0000000",
        "num_accounts": 32,
        "flags": {
          "auth_required": false,
          "auth_revocable": false
        }
      },
      {
        "_links": {
          "toml": {
            "href": "https://www.diamnet.org/.well-known/diamnet.toml"
          }
        },
        "asset_type": "credit_alphanum4",
        "asset_code": "USD",
        "asset_issuer": "GBAUUA74H4XOQYRSOW2RZUA4QL5PB37U3JS5NE3RTB2ELJVMIF5RLMAG",
        "paging_token": "USD_GBAUUA74H4XOQYRSOW2RZUA4QL5PB37U3JS5NE3RTB2ELJVMIF5RLMAG_credit_alphanum4",
        "amount": "1000000000.0000000",
        "num_accounts": 91547871,
        "flags": {
          "auth_required": false,
          "auth_revocable": false
        }
      }
    ]
  }
}
```

## Possible Errors

- The [standard errors](../errors.md#Standard_Errors).
