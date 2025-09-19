#!/bin/bash

set -e

echo "Generating protobuf files..."

cd "$(dirname "$0")/.."

protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       internal/proto/metrics.proto

echo "Protobuf files generated successfully!"