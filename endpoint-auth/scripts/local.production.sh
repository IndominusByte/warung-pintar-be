#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

export BACKEND_IMAGE=warungpintar-go-auth
export BACKEND_IMAGE_TAG=production
export BACKEND_CONTAINER=warungpintar-go-auth-production
export BACKEND_HOST=warungpintar-go-auth.service
export BACKEND_STAGE=production

docker build -t "$BACKEND_IMAGE:$BACKEND_IMAGE_TAG" .
docker-compose -f ./manifest/docker-compose.production.yaml up -d --build
