#!/usr/bin/env bash

go run cmd/gojni/main.go >env.go
export JAVA_HOME=`/usr/libexec/java_home` go build
