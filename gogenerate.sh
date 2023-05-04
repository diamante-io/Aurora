#! /bin/bash
set -e

# Check if go-bindata is installed, if not install it.
command -v go-bindata >/dev/null 2>&1 || (
  dir=$(mktemp -d)
  pushd $dir
  go mod init tool
  go get github.com/kevinburke/go-bindata/go-bindata@v3.18.0+incompatible
  popd
)

printf "Running go generate...\n"
go generate ./...

printf "Running make gxdr/xdr_generated.go...\n"
make gxdr/xdr_generated.go

printf "Checking for no diff...\n"
git diff --exit-code || (echo "Files changed after running go generate. Run go generate ./... locally and update generated files." && exit 1)
