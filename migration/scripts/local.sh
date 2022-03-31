#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

# db migration
export DB_MIGRATION_IMAGE=warung-pintar-db-migration
export DB_MIGRATION_IMAGE_TAG=development
export DB_MIGRATION_CONTAINER=warung-pintar-db-migration
export DB_MIGRATION_HOST=warung-pintar-db-migration.service
export BACKEND_STAGE=development

docker build -t "$DB_MIGRATION_IMAGE:$DB_MIGRATION_IMAGE_TAG" .
docker-compose -f ./manifest/docker-compose.yaml up -d --build
