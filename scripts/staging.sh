#!/bin/bash

export SLUG=ghcr.io/awakari/interests
export VERSION=latest
docker tag awakari/interests "${SLUG}":"${VERSION}"
docker push "${SLUG}":"${VERSION}"
