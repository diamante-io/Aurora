#! /usr/bin/env bash
set -e

# This scripts rebuilds the latest.sql file included in the schema package.
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GOTOP="$( cd "$DIR/../../../../../../../.." && pwd )"
go generate github.com/diamnet/go/services/aurora/internal/db2/schema
go install github.com/diamnet/go/services/aurora
dropdb aurora_schema --if-exists
createdb aurora_schema
DATABASE_URL=postgres://localhost/aurora_schema?sslmode=disable $GOTOP/bin/aurora db migrate up

DUMP_OPTS="--schema=public --no-owner --no-acl --inserts"
LATEST_PATH="$DIR/../db2/schema/latest.sql"
BLANK_PATH="$DIR/../test/scenarios/blank-aurora.sql"

pg_dump postgres://localhost/aurora_schema?sslmode=disable $DUMP_OPTS \
  | sed '/SET idle_in_transaction_session_timeout/d'  \
  | sed '/SET row_security/d' \
  > $LATEST_PATH
pg_dump postgres://localhost/aurora_schema?sslmode=disable \
  --clean --if-exists $DUMP_OPTS \
  | sed '/SET idle_in_transaction_session_timeout/d'  \
  | sed '/SET row_security/d' \
  > $BLANK_PATH

go generate github.com/diamnet/go/services/aurora/internal/db2/schema
go generate github.com/diamnet/go/services/aurora/internal/test
go install github.com/diamnet/go/services/aurora
