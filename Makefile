APP_API = ni-storage
.PHONY:all

PHONY:docker-clean
docker-clean:
	docker system prune -f --volumes

PHONY:docker-logs
docker-logs:
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

PHONY:build
build:
	go build -o ./bin/$(APP_API) ./cmd/api

PHONY:start
start:
	$(CURDIR)/bin/$(APP_API)

PHONY:test
test:
	go test -v -race -bench=. ./...

PHONY:release
release:
	docker build -f Dockerfile.release -t release-build:latest . && \
		docker run -it -v $(PWD)/bin/:/data/ release-build:latest cp -r /release/ /data