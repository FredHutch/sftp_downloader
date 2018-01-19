#!/bin/bash

set -e

go get github.com/golang/mock/gomock
go get github.com/golang/mock/mockgen

go generate ./...

mkdir -p builds

GOOS=linux   GOARCH=amd64 go build -o builds/sftp_downloader_linux
GOOS=windows GOARCH=amd64 go build -o builds/sftp_downloader_x64.exe
GOOS=darwin  GOARCH=amd64 go build -o builds/ftp_downloader_mac
