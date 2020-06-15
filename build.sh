#!/usr/bin/env bash

go run ./cmd/gen >env.go
includejni go build
