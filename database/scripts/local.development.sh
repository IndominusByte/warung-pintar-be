#!/bin/bash

export COMPOSE_IGNORE_ORPHANS=True

# postgresql
export POSTGRESQL_IMAGE=warungpintar-postgresql
export POSTGRESQL_IMAGE_TAG=development
export POSTGRESQL_CONTAINER=warungpintar-postgresql-development
export POSTGRESQL_HOST=warungpintar-postgresql.service
export POSTGRESQL_USER=warungpintardev
export POSTGRESQL_PASSWORD=inisecret
export POSTGRESQL_DB=warungpintar
export POSTGRESQL_TIME_ZONE=Asia/Kuala_Lumpur
docker build -t "$POSTGRESQL_IMAGE:$POSTGRESQL_IMAGE_TAG" -f ./manifest-docker/Dockerfile.postgresql ./manifest-docker

# redis
export REDIS_IMAGE=warungpintar-redis
export REDIS_IMAGE_TAG=development
export REDIS_CONTAINER=warungpintar-redis-development
export REDIS_HOST=warungpintar-redis.service
docker build -t "$REDIS_IMAGE:$REDIS_IMAGE_TAG" -f ./manifest-docker/Dockerfile.redis ./manifest-docker

# pgadmin
export PGADMIN_IMAGE=warungpintar-pgadmin
export PGADMIN_IMAGE_TAG=development
export PGADMIN_CONTAINER=warungpintar-pgadmin-development
export PGADMIN_HOST=warungpintar-pgadmin.service
export PGADMIN_EMAIL=admin@warungpintar.co.id
export PGADMIN_PASSWORD=inisecret
docker build -t "$PGADMIN_IMAGE:$PGADMIN_IMAGE_TAG" -f ./manifest-docker/Dockerfile.pgadmin ./manifest-docker

# pgbackups
export PGBACKUPS_IMAGE=warungpintar-pgbackups
export PGBACKUPS_IMAGE_TAG=development
export PGBACKUPS_CONTAINER=warungpintar-pgbackups-development
export PGBACKUPS_HOST=warungpintar-pgbackups.service
docker build -t "$PGBACKUPS_IMAGE:$PGBACKUPS_IMAGE_TAG" -f ./manifest-docker/Dockerfile.pgbackups ./manifest-docker

docker-compose -f ./manifest/docker-compose.development.yaml up -d --build
