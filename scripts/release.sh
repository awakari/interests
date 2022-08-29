#!/bin/bash

export SLUG=ghcr.io/cloud-messaging/subscriptions
export VERSION=$(./scripts/version.sh)
docker tag cloud-messaging/patterns "${SLUG}":"${VERSION}"
docker tag cloud-messaging/patterns "${SLUG}":latest
docker push "${SLUG}":"${VERSION}"
docker push "${SLUG}":latest
