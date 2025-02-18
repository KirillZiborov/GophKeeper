#!/bin/bash
#
# Script for building GophKeeper client.

set -e

build_version=$1
if [[ -z "$build_version" ]]; then
  echo "usage: $0 <build-version>"
  exit 1
fi

build_date=$(date +%F\ %H:%M:%S)

# Build info flags
ldflags="-X 'main.buildVersion=${build_version}' -X 'main.buildDate=${build_date}'"

package_name=gophkeeper
build_folder=./dist

echo "Building for linux/amd64..."
GOOS=linux GOARCH=amd64 go build \
    -ldflags "${ldflags}" \
    -o ${build_folder}/${package_name}-linux-amd64 \
    ./cmd/client

echo "Building for windows/amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build \
    -ldflags "${ldflags}" \
    -o ${build_folder}/${package_name}-windows-amd64.exe \
    ./cmd/client

echo "Building for darwin/amd64..."
GOOS=darwin GOARCH=amd64 CGO_ENABLED=1 go build \
    -ldflags "${ldflags}" \
    -o ${build_folder}/${package_name}-darwin-amd64 \
    ./cmd/client

echo "Building for darwin/arm64..."
GOOS=darwin GOARCH=arm64 CGO_ENABLED=1 go build \
    -ldflags "${ldflags}" \
    -o ${build_folder}/${package_name}-darwin-arm64 \
    ./cmd/client

echo "Finished build"