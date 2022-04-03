build:
	go build -v -o bin/http cmd/http/*.go

run: build
	bin/http

watch:
	reflex -s -r "\.(go|json|html)$$" --decoration=none make run

test:
	docker exec -it warungpintar-go-product-development go test ./tests/... -v

log-dev:
	docker logs -f --tail 10 warungpintar-go-product-development

log-prod:
	docker logs -f --tail 10 warungpintar-go-product-production

dev:
	bash ./scripts/local.development.sh

prod:
	bash ./scripts/local.production.sh
