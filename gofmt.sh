#! /bin/bash
set -e

printf "Running gofmt checks...\n"
OUTPUT=$(ls -d */ \
  | egrep -v '^vendor|^docs' \
  | xargs -I {} -P 4 gofmt -d {})

if [[ $OUTPUT ]]; then
  printf "gofmt found unformatted files:\n\n"
  echo "$OUTPUT"
  exit 1
fi
