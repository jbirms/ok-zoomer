#!/bin/sh
#GOOS=linux GOARCH=amd64 go build -o app core/*.go
CGO_ENABLED=1 xgo --targets=linux/amd64 --out app core/
