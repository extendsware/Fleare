#!/bin/bash

# Variables
PROJECT_NAME="fleare"
RELEASE_DIR="releases"
PLATFORMS=("linux/amd64" "darwin/arm64")

set -e

VERSION=$(cat VERSION)
IFS='.' read -r VERSION_MAJOR VERSION_MINOR VERSION_PATCH <<< "$VERSION"

# Create release directory
mkdir -p $RELEASE_DIR

# Build for each platform
for PLATFORM in "${PLATFORMS[@]}"; do
    OS=$(echo $PLATFORM | cut -d'/' -f1)
    ARCH=$(echo $PLATFORM | cut -d'/' -f2)
    OUTPUT_NAME="$PROJECT_NAME-$VERSION_MAJOR-$VERSION_MINOR-$VERSION_PATCH-$OS-$ARCH"

    if [ $OS = "windows" ]; then
        OUTPUT_NAME+='.exe'
    fi

    echo "Building for $OS/$ARCH..: $OUTPUT_NAME"
    env GOOS=$OS GOARCH=$ARCH go build -ldflags="-X 'main.VersionMajor=$VERSION_MAJOR' \
        -X 'main.VersionMinor=$VERSION_MINOR' \
        -X 'main.VersionPatch=$VERSION_PATCH' \
        -X 'main.Platform=$OS/$ARCH'" \
        -o $RELEASE_DIR/$OS-$ARCH/$OUTPUT_NAME

    if [ $? -ne 0 ]; then
        echo "An error occurred while building for $OS/$ARCH. Aborting."
        exit 1
    fi

    env GOOS=$OS GOARCH=$ARCH go build -o $RELEASE_DIR/$OS-$ARCH/$PROJECT_NAME-cli clients/cli/main.go
    if [ $? -ne 0 ]; then
        echo "An error occurred while building for $OS/$ARCH. Aborting."
        exit 1
    fi

    cp "install-$OS-$ARCH.sh" $RELEASE_DIR/$OS-$ARCH/install_fleare.sh

    xattr -cr $RELEASE_DIR/$OS-$ARCH
    # Optionally, compress the binary
    tar --disable-copyfile --exclude='*.tar.gz' --exclude='__MACOSX' --exclude='.DS_Store' -czf $RELEASE_DIR/$OS-$ARCH/$OUTPUT_NAME.tar.gz -C $RELEASE_DIR/$OS-$ARCH .

    rm $RELEASE_DIR/$OS-$ARCH/$OUTPUT_NAME
    rm $RELEASE_DIR/$OS-$ARCH/$PROJECT_NAME-cli
    rm $RELEASE_DIR/$OS-$ARCH/install_fleare.sh
done

echo "Building docker image...Tag: $VERSION"
docker build --no-cache --build-arg APP_VERSION=$VERSION_MAJOR-$VERSION_MINOR-$VERSION_PATCH  -t extendsware/fleare:$VERSION .

echo "Building docker image...Tag: latest"
docker build --no-cache --build-arg APP_VERSION=$VERSION_MAJOR-$VERSION_MINOR-$VERSION_PATCH -t extendsware/fleare .

echo "Build complete. Release files are located in the $RELEASE_DIR directory."
