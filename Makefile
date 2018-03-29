NAME = multy

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"

all:  build run

all-with-deps: setup deps docker

all-docker:  setup deps
	cd cmd && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o $(NAME) .

run:
	cd cmd && ./$(NAME) && ../

setup:
	go get -u github.com/kardianos/govendor

deps:
	govendor sync


	
build:
	cd node-streamer/btc/ && protoc --go_out=plugins=grpc:. *.proto && cd ../../cmd/ && go build $(LD_OPTS) -o $(NAME) . && cd -

race:
	cd node-streamer/btc/ && protoc --go_out=plugins=grpc:. *.proto && cd ../../cmd/ && go build $(LD_OPTS) -o $(NAME) -race . && cd -

# Show to-do items per file.
todo:
	@grep \
		--exclude-dir=vendor \
		--exclude-dir=node_modules \
		--exclude=Makefile \
		--text \
		--color \
		-nRo -E ' TODO:.*|SkipNow|nolint:.*' .
.PHONY: todo

dist:
	cd node-streamer/eth && protoc --go_out=plugins=grpc:. *.proto &&cd ../../cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o $(NAME) .

test:
	cd cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o test .

stage:
	cd cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o stage .

proto:
	cd node-streamer/btc && protoc --go_out=plugins=grpc:. *.proto && cd .. && cd eth/ && protoc --go_out=plugins=grpc:. *.proto