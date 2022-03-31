build:
	go build -v -o bin/migration *.go

run: build
	bin/migration

dev:
	bash ./scripts/local.sh

log-dev:
	docker logs -f --tail 10 warung-pintar-db-migration
