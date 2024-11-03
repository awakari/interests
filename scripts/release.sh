#!/bin/bash

export SLUG=ghcr.io/awakari/interests
export VERSION=$(git describe --tags --abbrev=0 | cut -c 2-)
echo "Releasing version: $VERSION"
docker tag awakari/interests "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
