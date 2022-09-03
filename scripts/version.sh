#!/bin/sh

sed -n "s/^\s*version\s*=\s*\"\(\S*\)\".*$/\1/p" main.go
