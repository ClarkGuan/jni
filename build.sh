#!/usr/bin/env bash

go run cmd/gojni/main.go -p demo >env.go
go build
