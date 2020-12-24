---
title: Fee Stats
clientData:
  laboratoryUrl:
---

This endpoint gives useful information about per-operation fee stats in the last 5 ledgers. It can be used to
predict a fee set on the transaction that will be submitted to the network.

Note: This endpoint was originally `/operation_fee_stats`, which is subject for removal in `0.18.0`.

## Request

```
GET /fee_stats
```

### curl Example Request

```sh
curl "https://aurora-testnet.diamnet.org/fee_stats"
```

## Response

Response contains the following fields:

| Field | |
| - | - |
| last_ledger | Last ledger sequence number |
| last_ledger_base_fee | Base fee as defined in the last ledger |
| ledger_capacity_usage | Average capacity usage in the last 5 ledgers. (0 is no usage, 1.0 is completely full ledgers) |
| min_accepted_fee | Minimum accepted fee in the last 5 ledger. |
| mode_accepted_fee | Mode accepted fee in the last 5 ledger. |
| p10_accepted_fee | 10th percentile accepted fee in the last 5 ledger. |
| p20_accepted_fee | 20th percentile accepted fee in the last 5 ledger. |
| p30_accepted_fee | 30th percentile accepted fee in the last 5 ledger. |
| p40_accepted_fee | 40th percentile accepted fee in the last 5 ledger. |
| p50_accepted_fee | 50th percentile accepted fee in the last 5 ledger. |
| p60_accepted_fee | 60th percentile accepted fee in the last 5 ledger. |
| p70_accepted_fee | 70th percentile accepted fee in the last 5 ledger. |
| p80_accepted_fee | 80th percentile accepted fee in the last 5 ledger. |
| p90_accepted_fee | 90th percentile accepted fee in the last 5 ledger. |
| p95_accepted_fee | 95th percentile accepted fee in the last 5 ledger. |
| p99_accepted_fee | 99th percentile accepted fee in the last 5 ledger. |

### Example Response

```json
{
  "last_ledger": "22606298",
  "last_ledger_base_fee": "100",
  "ledger_capacity_usage": "0.97",
  "min_accepted_fee": "100",
  "mode_accepted_fee": "250",
  "p10_accepted_fee": "100",
  "p20_accepted_fee": "100",
  "p30_accepted_fee": "250",
  "p40_accepted_fee": "250",
  "p50_accepted_fee": "250",
  "p60_accepted_fee": "1210",
  "p70_accepted_fee": "1221",
  "p80_accepted_fee": "1225",
  "p90_accepted_fee": "1225",
  "p95_accepted_fee": "1225",
  "p99_accepted_fee": "8000"
}
```

## Possible Errors

- The [standard errors](../errors.md#Standard_Errors).
