#! /bin/bash
set -e

if [ -z ${LATEST_LEDGER+x} ]; then
    # Get latest ledger
    echo "Getting latest checkpoint ledger..."
    if [ -z ${TESTNET+x} ]; then
        export LATEST_LEDGER=`curl -s http://history.diamnet.org/prd/core-live/core_live_001/.well-known/diamnet-history.json | jq -r '.currentLedger'`
    else
        export LATEST_LEDGER=`curl -s http://history.diamnet.org/prd/core-testnet/core_testnet_001/.well-known/diamnet-history.json | jq -r '.currentLedger'`
    fi
    echo "Latest ledger: $LATEST_LEDGER"
fi

# Dump state using Golang
if [ -z ${TESTNET+x} ]; then
    echo "Dumping pubnet state using ingest..."
    go run ./main.go
else
    echo "Dumping testnet state using ingest..."
    go run ./main.go --testnet
fi
echo "State dumped..."

# Catchup core
if [ -z ${TESTNET+x} ]; then
    echo "Catch up from pubnet"
    diamnet-core --conf ./diamnet-core.cfg catchup $LATEST_LEDGER/1
else
    echo "Catch up from testnet"
    diamnet-core --conf ./diamnet-core-testnet.cfg catchup $LATEST_LEDGER/1
fi

echo "Dumping state from diamnet-core..."
./dump_core_db.sh
echo "State dumped..."

echo "Comparing state dumps..."
./diff_test.sh
