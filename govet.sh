#! /bin/bash
set -e

# TODO: 1) Build and install shadow analyzer for go vet in 1.12+
# See https://github.com/golang/go/issues/29260
# https://golang.org/doc/go1.12#vet
# For now we just skip the shadow checking for Go 1.12+
# TODO: 2) Not triggering on shadowed vars (output not stdout?)
# TODO: 3) Fix syntax for Go 1.12+ (go tool vet -> go vet, but fails to find packages?)
if [[ $GOLANG_VERSION = 1.9.* ]] || [[ $GOLANG_VERSION = 1.10.* ]] || [[ $GOLANG_VERSION = 1.11.* ]]; then
    echo "Running go vet checks..."
OUTPUT=$(ls -d */ \
  | egrep -v '^vendor|^docs' \
  | xargs -I {} -P 4 go tool vet -all -composites=false -unreachable=false -tests=false -shadow {})
else
    echo "Skipping go vet checks for this version of Go..."
fi


if [[ $OUTPUT ]]; then
  printf "govet found some issues:\n\n"
  echo "$OUTPUT"
  exit 1
fi
