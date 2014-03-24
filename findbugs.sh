#!/bin/sh
set -e

echo "Vetting project."
go vet || exit 1
echo "...ok"

echo "Checking for missing error handling."
errcheck `go list`/... || exit 2
echo "...ok"

echo "Running tests."
go test ./... || exit 3
echo "...ok"

# let go for now. golint fucks up because of bin2go stupid filenames
# echo "Linting project."
# golint **/*.go  || exit 3
# echo "...ok"
