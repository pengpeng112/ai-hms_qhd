#!/bin/bash
set -e
echo "=== Backend Verification ==="
echo "--- go fmt ---"
UNFMT=$(gofmt -l .)
if [ -n "$UNFMT" ]; then
  echo "FAIL: Unformatted files:"
  echo "$UNFMT"
  exit 1
fi
echo "PASS: go fmt"

echo "--- go vet ---"
go vet ./...
echo "PASS: go vet"

echo "--- go build ---"
go build ./cmd/server
echo "PASS: go build"

echo "=== All backend checks passed ==="
