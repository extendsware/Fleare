#!/bin/bash

# Variables
PROJECT_NAME="fleare"

set -e

VERSION=$(cat VERSION)
IFS='.' read -r VERSION_MAJOR VERSION_MINOR VERSION_PATCH BUILD_DATE<<< "$VERSION"

tee "commander/version.go" > /dev/null <<EOL
package commander

var (
	ProjectName = "$PROJECT_NAME"
	Version     = "v$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH"
	BuildDate   = "$BUILD_DATE"
)
EOL

echo "Building docker image...Tag: $VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH"
docker build --no-cache --build-arg APP_VERSION=$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH  --progress=plain -t extendsware/fleare:$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH .

echo "Building docker image...Tag: latest"
docker build --no-cache --build-arg APP_VERSION=$VERSION_MAJOR.$VERSION_MINOR.$VERSION_PATCH -t extendsware/fleare:latest .

echo "Build complete. Release files are located in the $RELEASE_DIR directory."
