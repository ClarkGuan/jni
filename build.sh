#!/usr/bin/env bash

go run cmd/gojni/main.go >env.go
JAVA_HOME=`/usr/libexec/java_home` CGO_CFLAGS="-I$JAVA_HOME/include -I$JAVA_HOME/include/darwin" go build
