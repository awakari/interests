#!/bin/bash

export SLUG=ghcr.io/meandros-messaging/subscriptions
export VERSION=$(./scripts/version.sh)
docker tag meandros-messaging/subscriptions "${SLUG}":"${VERSION}"
docker tag meandros-messaging/subscriptions "${SLUG}":latest
docker push "${SLUG}":"${VERSION}"
docker push "${SLUG}":latest
