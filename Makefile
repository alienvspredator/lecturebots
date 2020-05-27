GIT_COMMIT := $(shell git rev-list -1 HEAD)

all: bot

bot:
	go build -o ./build/bot -ldflags "-X main.GitCommit=$(GIT_COMMIT)" ./cmd/bot

clean:
	rm -rf ./build/*
