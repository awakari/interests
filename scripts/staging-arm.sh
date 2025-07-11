#!/bin/bash

export SLUG=ghcr.io/awakari/interests-arm64
export VERSION=latest
docker tag awakari/interests "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
