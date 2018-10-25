#!/usr/bin/env bash

go run cmd/gojni/main.go >env.go
go build
