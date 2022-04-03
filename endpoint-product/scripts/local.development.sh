#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

export BACKEND_IMAGE=warungpintar-go-product
export BACKEND_IMAGE_TAG=development
export BACKEND_CONTAINER=warungpintar-go-product-development
export BACKEND_HOST=warungpintar-go-product.service
export BACKEND_STAGE=development

docker build -t "$BACKEND_IMAGE:$BACKEND_IMAGE_TAG" .
docker-compose -f ./manifest/docker-compose.development.yaml up -d --build
