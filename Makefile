MAKEFLAGS += --no-print-directory

all:
	@make packages
	@make server
	@make client
	@bin/gochat-server

packages:
	@go build ./pkg/parser

server:
	@go build -o bin/ ./cmd/gochat-server

client:
	@go build -o bin/ ./cmd/gochat-client

run:
	@bin/gochat-server