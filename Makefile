NAME = client

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"

all:  build run

all-with-deps: setup deps build

run: build
	cd cmd && ./$(NAME) && ../

memprofiler: build
	cd cmd && ./$(NAME) -memprofile mem.prof && ../

setup:
	go get -u github.com/kardianos/govendor

deps:
	govendor sync

build:
	cd cmd && go build $(LD_OPTS) -o $(NAME) . && cd -

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
	cd cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o $(NAME) .

test:
	cd cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o test .

stage:
	cd cmd/ && GOOS=linux GOARCH=amd64 go build $(LD_OPTS)  -o stage .

proto:
	cd ./node-streamer && protoc --go_out=plugins=grpc:. *.proto 