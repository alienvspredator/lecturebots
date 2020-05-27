GIT_COMMIT := $(shell git rev-list -1 HEAD)

all: bot

bot:
	go build -o ./build/bot -ldflags "-X main.gitCommit=$(GIT_COMMIT)" ./cmd/bot

clean:
	rm -rf ./build/*

container:
	docker-compose --env-file deployments/.env -f deployments/docker-compose.yaml up --build --abort-on-container-exit bot

test-db:
	docker-compose --env-file deployments/.env -f deployments/docker-compose.dev.yaml up -d postgres pgadmin
