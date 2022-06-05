#!/usr/bin/env bash
# First argument goes to ghr
# Second argument should be platform (windows, linux, etc)
# Next arguments are the architetures to build to

VERSION=$(git describe --tags --abbrev=0)
delete=$1
shift
platform=$1
shift
archs=$@

cd build/
mkdir -p release

wget https://github.com/tcnksm/ghr/releases/download/v0.14.0/ghr_v0.14.0_linux_amd64.tar.gz
tar -xvzf ghr_*.tar.gz
mv ghr_*_amd64 ghr

for arch in ${archs[@]}
do
    zip "${platform}_${arch}.zip" "$arch/blackbeard"
done
mv *.zip release/

echo "RELEASE VERSION $VERSION"
./ghr/ghr -t ${GITHUB_TOKEN} -u ${CIRCLE_PROJECT_USERNAME} -r ${CIRCLE_PROJECT_REPONAME} -c ${CIRCLE_SHA1} $delete ${VERSION} ./release/
