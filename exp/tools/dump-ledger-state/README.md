# dump-ledger-state

This tool dumps the state from history archive buckets to 4 separate files:
* accounts.csv
* accountdata.csv
* offers.csv
* trustlines.csv

It's primary use is to test `SingleLedgerStateReader`. To test it:
1. Run `dump-ledger-state`.
2. Sync diamnet-core to the same checkpoint: `diamnet-core catchup [ledger]/1`.
3. Dump diamnet-core DB by using `dump_core_db.sh` script.
4. Diff results by using `diff_test.sh` script.
