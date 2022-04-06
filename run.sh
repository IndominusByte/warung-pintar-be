#!/bin/bash
docker network create warungpintar-environment-development
echo "====== CREATE DB ======"
cd database
make dev
cd ..
echo "====== MIGRATE DB ======"
cd migration
make dev
cd ..
echo "====== RUNNING REST AUTH ======"
cd endpoint-auth
make dev
echo "====== RUNNING TEST AUTH ======"
make test
cd ..
echo "====== RUNNING REST PRODUCT ======"
cd endpoint-product
make dev
echo "====== RUNNING TEST PRODUCT ======"
make test
cd ..
echo "====== RUNNING REST TRANSACTION ======"
cd endpoint-transaction
make dev
echo "====== RUNNING TEST TRANSACTION ======"
make test
cd ..
