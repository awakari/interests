#!/bin/bash

export SLUG=ghcr.io/meandros-messaging/subscriptions
export VERSION=$(git describe --tags --abbrev=0)
echo "Releasing version: $VERSION"
docker tag meandros-messaging/subscriptions "${SLUG}":"${VERSION}"
docker tag meandros-messaging/subscriptions "${SLUG}":latest
docker push "${SLUG}":"${VERSION}"
docker push "${SLUG}":latest
