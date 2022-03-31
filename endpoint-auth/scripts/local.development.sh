#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

export BACKEND_IMAGE=warungpintar-go-auth
export BACKEND_IMAGE_TAG=development
export BACKEND_CONTAINER=warungpintar-go-auth-development
export BACKEND_HOST=warungpintar-go-auth.service
export BACKEND_STAGE=development

docker build -t "$BACKEND_IMAGE:$BACKEND_IMAGE_TAG" .
docker-compose -f ./manifest/docker-compose.development.yaml up -d --build
