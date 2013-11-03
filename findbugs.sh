#!/bin/sh
set -e

echo "Vetting project."
go vet
echo "...ok"

echo "Checking for missing error handling."
errcheck `go list`/...
echo "...ok"

echo "Linting project."
golint **/*.go
echo "...ok"
