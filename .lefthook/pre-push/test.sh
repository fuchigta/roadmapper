#!/usr/bin/env bash
set -e
echo "go test -race ./..."
go test -race ./...
