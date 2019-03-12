APP_API = ni-storage
.PHONY:all

PHONY:docker-clean
docker-clean:
	docker system prune -f --volumes

PHONY:logs
logs:
	docker-compose logs -f api

PHONY:docker-start
docker-start:
	docker-compose stop api
	docker-compose rm -f api
	docker-compose up -d api

PHONY:docker-stop
docker-stop:
	docker-compose stop

PHONY:docker-build
docker-build:
	docker build --rm --tag filatovw/$(APP_API) .

PHONY: clean
clean:
	rm bin/$(APP_API)

PHONY:build
build:
	go build -o ./bin/$(APP_API) ./cmd/api

PHONY:start
start:
	./scripts/start.sh

PHONY:stop
stop:
	./scripts/stop.sh
