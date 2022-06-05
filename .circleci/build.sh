#!/usr/bin/env bash

version=$(git describe --tags --abbrev=0)
version=${version/[a-zA-Z]/}

archs=(amd64 arm64 386)

for arch in ${archs[@]}
do
    env GOOS=$1 GOARCH=${arch} CGO_ENABLED=1 go build -ldflags="-X 'main.Version=$version' -X 'main.BuildDate=$(date -u)'" -o prepnode_${arch}

done
