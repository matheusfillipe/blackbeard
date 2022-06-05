#!/usr/bin/env bash

version=$(git describe --tags --abbrev=0)
version=${version/[a-zA-Z]/}


go build -ldflags="-X 'main.Version=$version' -X 'main.BuildDate=$(date -u)'"
