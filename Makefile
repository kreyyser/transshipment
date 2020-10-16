GO_SERVICES = restgateway ports


.PHONY: test
test:
	go test ./... -v

.PHONY: lint
lint: lint-ports lint-restgateway

.PHONY: lint-ports
lint-ports:
	cd ports && make lint

.PHONY: lint-restgateway
lint-restgateway:
	cd restgateway && make lint

.PHONY: all
buil-all: ports restgateway

.PHONY: ports
ports:
	cd ports && make ports

.PHONY: restgateway
restgateway:
	cd restgateway && make restgateway

.PHONY: spin
spin:
	COMPOSE_PROJECT_NAME=transhipment docker-compose up -d

.PHONY: start
start: buil-all spin

.PHONY: stop
stop:
	COMPOSE_PROJECT_NAME=transhipment docker-compose down

.PHONY: proto
proto:
	cd pb && ./protobuf.sh

