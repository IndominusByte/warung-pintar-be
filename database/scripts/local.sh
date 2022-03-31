#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

# postgresql
export POSTGRESQL_IMAGE=warung-pintar-postgresql
export POSTGRESQL_IMAGE_TAG=development
export POSTGRESQL_CONTAINER=warung-pintar-postgresql
export POSTGRESQL_HOST=warung-pintar-postgresql.service
export POSTGRESQL_USER=warung-pintardev
export POSTGRESQL_PASSWORD=inisecret
export POSTGRESQL_DB=warung-pintar
export POSTGRESQL_TIME_ZONE=Asia/Kuala_Lumpur
docker build -t "$POSTGRESQL_IMAGE:$POSTGRESQL_IMAGE_TAG" -f ./manifest-docker/Dockerfile.postgresql ./manifest-docker

# redis
export REDIS_IMAGE=warung-pintar-redis
export REDIS_IMAGE_TAG=development
export REDIS_CONTAINER=warung-pintar-redis
export REDIS_HOST=warung-pintar-redis.service
docker build -t "$REDIS_IMAGE:$REDIS_IMAGE_TAG" -f ./manifest-docker/Dockerfile.redis ./manifest-docker

# pgadmin
export PGADMIN_IMAGE=warung-pintar-pgadmin
export PGADMIN_IMAGE_TAG=development
export PGADMIN_CONTAINER=warung-pintar-pgadmin
export PGADMIN_HOST=warung-pintar-pgadmin.service
export PGADMIN_EMAIL=admin@warung-pintargroup.co.id
export PGADMIN_PASSWORD=inisecret
docker build -t "$PGADMIN_IMAGE:$PGADMIN_IMAGE_TAG" -f ./manifest-docker/Dockerfile.pgadmin ./manifest-docker

# pgbackups
export PGBACKUPS_IMAGE=warung-pintar-pgbackups
export PGBACKUPS_IMAGE_TAG=development
export PGBACKUPS_CONTAINER=warung-pintar-pgbackups
export PGBACKUPS_HOST=warung-pintar-pgbackups.service
docker build -t "$PGBACKUPS_IMAGE:$PGBACKUPS_IMAGE_TAG" -f ./manifest-docker/Dockerfile.pgbackups ./manifest-docker

docker-compose -f ./manifest/docker-compose.yaml up -d --build
