#! /usr/bin/env bash
set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOTOP="$( cd "$DIR/../../../../../../../.." && pwd )"
PACKAGES=$(find $GOTOP/src/github.com/hcnet/go/services/aurora/internal/test/scenarios -iname '*.rb' -not -name '_common_accounts.rb')
#PACKAGES=$(find $GOTOP/src/github.com/hcnet/go/services/aurora/internal/test/scenarios -iname 'failed_transactions.rb')

go install github.com/hcnet/go/services/aurora

dropdb hayashi_scenarios --if-exists
createdb hayashi_scenarios

export HCNET_CORE_DATABASE_URL="postgres://localhost/hayashi_scenarios?sslmode=disable"
export DATABASE_URL="postgres://localhost/aurora_scenarios?sslmode=disable"
export NETWORK_PASSPHRASE="Test SDF Network ; September 2015"
export HCNET_CORE_URL="http://localhost:8080"
export SKIP_CURSOR_UPDATE="true"
export INGEST_FAILED_TRANSACTIONS=true

# run all scenarios
for i in $PACKAGES; do
  echo $i
  CORE_SQL="${i%.rb}-core.sql"
  HORIZON_SQL="${i%.rb}-aurora.sql"
  scc -r $i --allow-failed-transactions --dump-root-db > $CORE_SQL

  # load the core scenario
  psql $HCNET_CORE_DATABASE_URL < $CORE_SQL

  # recreate aurora dbs
  dropdb aurora_scenarios --if-exists
  createdb aurora_scenarios

  # import the core data into aurora
  $GOTOP/bin/aurora db init
  $GOTOP/bin/aurora db init-asset-stats
  $GOTOP/bin/aurora db rebase

  # write aurora data to sql file
  pg_dump $DATABASE_URL \
    --clean --if-exists --no-owner --no-acl --inserts \
    | sed '/SET idle_in_transaction_session_timeout/d' \
    | sed '/SET row_security/d' \
    > $HORIZON_SQL
done


# commit new sql files to bindata
go generate github.com/hcnet/go/services/aurora/internal/test/scenarios
# go test github.com/hcnet/go/services/aurora/internal/ingest
