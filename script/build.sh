#!/bin/bash

SERVICE="benzene"

# Take advantage of the cross-compilation feature of the Go language.
export GO111MODULE=on
export GOOS=linux
export GOARCH=amd64
export CGO_ENABLED=0

# Organize build artifacts and configuration files.
mkdir -p output/config
go build -o output/${SERVICE}
cp -r config/* output/config

# Build the docker image to deployment.
export DOCKER_BUILDKIT=0  
docker build -t benzene:latest .