#!/usr/bin/env bash

usage() {
    echo "USAGE: .scripts/release.sh [version] [msg...]"
    exit 1
}

REVISION=$(git rev-parse HEAD)
GIT_TAG=$(git name-rev --tags --name-only $REVISION)
if [ "$GIT_TAG" = "" ]; then
    GIT_TAG="devel"
fi


VERSION=$1
if [ "$VERSION" = "" ]; then
    echo "Need to specify a version! Perhaps '$GIT_TAG'?"
    usage
fi

set -u -e

echo "Brog version $VERSION"
rm -rf /tmp/brog_build/

mkdir -p /tmp/brog_build/linux
GOOS=linux godep go build -ldflags "-X main.version=$VERSION" -o /tmp/brog_build/linux/brog ./
pushd /tmp/brog_build/linux/
tar cvzf /tmp/brog_build/brog_linux.tar.gz brog
popd

mkdir -p /tmp/brog_build/darwin
GOOS=darwin godep go build -ldflags "-X main.version=$VERSION" -o /tmp/brog_build/darwin/brog ./
pushd /tmp/brog_build/darwin/
tar cvzf /tmp/brog_build/brog_darwin.tar.gz brog
popd

temple file < scripts/README.tmpl.md > README.md -var "version=$VERSION"
git add README.md
git commit -m 'release bump'

hub release create \
    -a /tmp/brog_build/brog_linux.tar.gz \
    -a /tmp/brog_build/brog_darwin.tar.gz \
    $VERSION

git push origin master
