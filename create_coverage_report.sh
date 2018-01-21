#!/usr/bin/env bash

set -e
# clear coverage.out
rm -f coverage.out cover-profile.out

cp /dev/null coverage.out

for d in $(go list ./... | grep -v vendor); do
    go test -race -coverprofile=profile.out -covermode=atomic $d
    if [ -f profile.out ]; then
        cat profile.out >> coverage.out
        rm profile.out
    fi
done

# remove duplicate lines (mode: count is present in each package)
awk '!a[$0]++' coverage.out > cover-profile.out
go tool cover -html=cover-profile.out -o coverage.html
# rm coverage.out cover-profile.out

# echo "total:"

go tool cover -func=cover-profile.out
