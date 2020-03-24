---
title: Transaction
---

**Transactions** are the basic unit of change in the HcNet Network.

A transaction is a grouping of [operations](./operation.md).

To learn more about the concept of transactions in the HcNet network, take a look at the [HcNet transactions concept guide](https://www.hcnet.org/developers/learn/concepts/transactions.html).

## Attributes

|    Attribute     |  Type  |                                                                                                                                |
| ---------------- | ------ | ------------------------------------------------------------------------------------------------------------------------------ |
| id               | string | The canonical id of this transaction, suitable for use as the :id parameter for url templates that require a transaction's ID. |
| paging_token     | string | A [paging token](./page.md) suitable for use as the `cursor` parameter to transaction collection resources.                   |
| successful       | bool | *From 0.17.0* Indicates if transaction was successful or not.       |
| hash             | string | A hex-encoded SHA-256 hash of the transaction's [XDR](../../learn/xdr.md)-encoded form.                                                              |
| ledger           | number | Sequence number of the ledger in which this transaction was applied.       |
| created_at       | ISO8601 string | |
| source_account   | string |                                                                                                                                |
| source_account_sequence | string |                                                                                                                                |
| fee_paid         | number | The fee paid by the source account of this transaction when the transaction was applied to the ledger.                         |
| operation_count  | number | The number of operations that are contained within this transaction.                                                           |
| envelope_xdr     | string | A base64 encoded string of the raw `TransactionEnvelope` xdr struct for this transaction                                       |
| result_xdr       | string | A base64 encoded string of the raw `TransactionResult` xdr struct for this transaction                                     |
| result_meta_xdr  | string | A base64 encoded string of the raw `TransactionMeta` xdr struct for this transaction                                           |
| fee_meta_xdr     | string | A base64 encoded string of the raw `LedgerEntryChanges` xdr struct produced by taking fees for this transaction.                                           |
| memo_type        | string | |
| memo             | string | |
| signatures | string[] | An array of signatures used to sign this transaction |
| valid_after | ISO8601 string | |
| valid_before | ISO8601 string | |

## Links

|                   rel                    |                                           Example                                           |                             Description                          |
| ---------------------------------------- | ------------------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| self | `https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f` |  |
| account | `https://aurora-testnet.hcnet.org/accounts/GCDLRUXOD6KA53G5ILL435TZAISNLPS4EKIHSOVY3MVD3DVJ333NO4DT` | The source [account](../accounts-single.md) for this transaction. |
| ledger | `https://aurora-testnet.hcnet.org/ledgers/2352988` | The [ledger](../ledgers-single.md) in which this transaction was applied. |
| operations | `https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f/operations{?cursor,limit,order}"` | [Operations](../operations-for-transaction.md) included in this transaction. |
| effects | `https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f/effects{?cursor,limit,order}"` | [Effects](../effects-for-transaction.md) which resulted by operations in this transaction. |
| precedes | `https://aurora-testnet.hcnet.org/transactions?order=asc&cursor=10106006507900928` | A collection of transactions that occur after this transaction. |
| succeeds | `https://aurora-testnet.hcnet.org/transactions?order=desc&cursor=10106006507900928` | A collection of transactions that occur before this transaction. |

## Example

```json
{
  "_links": {
    "self": {
      "href": "https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f"
    },
    "account": {
      "href": "https://aurora-testnet.hcnet.org/accounts/GCDLRUXOD6KA53G5ILL435TZAISNLPS4EKIHSOVY3MVD3DVJ333NO4DT"
    },
    "ledger": {
      "href": "https://aurora-testnet.hcnet.org/ledgers/2352988"
    },
    "operations": {
      "href": "https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f/operations{?cursor,limit,order}",
      "templated": true
    },
    "effects": {
      "href": "https://aurora-testnet.hcnet.org/transactions/cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f/effects{?cursor,limit,order}",
      "templated": true
    },
    "precedes": {
      "href": "https://aurora-testnet.hcnet.org/transactions?order=asc&cursor=10106006507900928"
    },
    "succeeds": {
      "href": "https://aurora-testnet.hcnet.org/transactions?order=desc&cursor=10106006507900928"
    }
  },
  "id": "cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f",
  "paging_token": "10106006507900928",
  "successful": true,
  "hash": "cb9a25394acb6fe0d1d9bdea5afc01cafe2c6fde59a96ddceb2564a65780a81f",
  "ledger": 2352988,
  "created_at": "2019-02-21T21:44:13Z",
  "source_account": "GCDLRUXOD6KA53G5ILL435TZAISNLPS4EKIHSOVY3MVD3DVJ333NO4DT",
  "source_account_sequence": "10105916313567234",
  "fee_paid": 100,
  "operation_count": 1,
  "envelope_xdr": "AAAAAIa40u4flA7s3ULXzfZ5AiTVvlwikHk6uNsqPY6p3vbXAAAAZAAj50cAAAACAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAABAAAAB2Fmc2RmYXMAAAAAAQAAAAAAAAABAAAAAIa40u4flA7s3ULXzfZ5AiTVvlwikHk6uNsqPY6p3vbXAAAAAAAAAAEqBfIAAAAAAAAAAAGp3vbXAAAAQKElK3CoNo1f8fWIGeJm98lw2AaFiyVVFhx3uFok0XVW3MHV9MubtEhfA+n1iLPrxmzHtHfmZsumWk+sOEQlSwI=",
  "result_xdr": "AAAAAAAAAGQAAAAAAAAAAQAAAAAAAAABAAAAAAAAAAA=",
  "result_meta_xdr": "AAAAAQAAAAIAAAADACPnXAAAAAAAAAAAhrjS7h+UDuzdQtfN9nkCJNW+XCKQeTq42yo9jqne9tcAAAAXSHbnOAAj50cAAAABAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABACPnXAAAAAAAAAAAhrjS7h+UDuzdQtfN9nkCJNW+XCKQeTq42yo9jqne9tcAAAAXSHbnOAAj50cAAAACAAAAAAAAAAAAAAAAAAAAAAEAAAAAAAAAAAAAAAAAAAAAAAABAAAAAA==",
  "fee_meta_xdr": "AAAAAgAAAAMAI+dTAAAAAAAAAACGuNLuH5QO7N1C1832eQIk1b5cIpB5OrjbKj2Oqd721wAAABdIduecACPnRwAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAEAI+dcAAAAAAAAAACGuNLuH5QO7N1C1832eQIk1b5cIpB5OrjbKj2Oqd721wAAABdIduc4ACPnRwAAAAEAAAAAAAAAAAAAAAAAAAAAAQAAAAAAAAAAAAAAAAAAAA==",
  "memo_type": "text",
  "memo": "afsdfas",
  "valid_after": "1970-01-01T00:00:00Z",
  "signatures": [
    "oSUrcKg2jV/x9YgZ4mb3yXDYBoWLJVUWHHe4WiTRdVbcwdX0y5u0SF8D6fWIs+vGbMe0d+Zmy6ZaT6w4RCVLAg=="
  ]
}
```

## Endpoints

|  Resource                |    Type    |    Resource URI Template             |
| ------------------------ | ---------- | ------------------------------------ |
| [All Transactions](../transactions-all.md)     | Collection | `/transactions` (`GET`) |
| [Post Transaction](../transactions-create.md)     | Action | `/transactions`  (`POST`) |
| [Transaction Details](../transactions-single.md)  | Single     | `/transactions/:id` |
| [Account Transactions](../transactions-for-account.md) | Collection | `/accounts/:account_id/transactions` |
| [Ledger Transactions](../transactions-for-ledger.md)  | Collection | `/ledgers/:ledger_id/transactions`   |


## Submitting transactions
To submit a new transaction to HcNet network, it must first be built and signed locally. Then you can submit a hex representation of your transaction’s [XDR](../../learn/xdr.md) to the `/transactions` endpoint. Read more about submitting transactions in [Post Transaction](../transactions-create.md) doc.
