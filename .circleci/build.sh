#!/usr/bin/env bash
# First argument should be platform (windows, linux, etc)
# Next arguments are the architetures to build to
# List architetures with `go tool dist list | grep linux`

version=$(git describe --tags --abbrev=0)
version=${version/[a-zA-Z]/}

platform=$1
shift
archs=$@

mkdir -p build

echo "Platform $platform"
for arch in ${archs[@]}
do
    echo "  Building $arch"
    env GOOS=${platform} GOARCH=${arch} CGO_ENABLED=1 go build -ldflags="-X 'main.Version=$version' -X 'main.BuildDate=$(date -u)'" -o "build/$arch/blackbeard"
done
