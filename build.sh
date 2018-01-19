#!/bin/bash

set -e

go get github.com/golang/mock/gomock
go get github.com/golang/mock/mockgen

go generate ./...

LDFLAGS="-X main.gitCommit=$(git rev-parse --short HEAD) -X main.gitBranch=$(git rev-parse --abbrev-ref HEAD)"

# echo "$LDFLAGS"

go install -ldflags "$LDFLAGS"

mkdir -p builds mocks

GOOS=linux   GOARCH=amd64 go build -ldflags "$LDFLAGS" -o builds/sftp_downloader_linux
GOOS=windows GOARCH=amd64 go build -ldflags "$LDFLAGS"  -o builds/sftp_downloader_x64.exe
GOOS=darwin  GOARCH=amd64 go build -ldflags "$LDFLAGS" -o builds/ftp_downloader_mac
