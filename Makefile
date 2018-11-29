NAME = multy

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
LASTTAG = $(shell git describe --tags --abbrev=0 --dirty)
GOPATH = $(shell echo "$$GOPATH")
LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.lasttag=${LASTTAG} -X main.buildtime=${BUILDTIME} -w "

# List of all binary targets we expect from make to produce
TARGETS=cmd/multy-back/multy-back cmd/ns-btc/ns-btc cmd/ns-eth/ns-eth

# List of all docker images to build and tag
DOCKER_IMAGES=multy-back multy-btc-node-service multy-eth-node-service
DOCKER_TAG?=latest

TARGET_OS=
TARGET_ARCH=

all: proto build test

all-with-deps: setup deps dist

# run:
# 	cd cmd && ./$(NAME) && ../

# memprofiler:
# 	cd $(GOPATH)/src/github.com/Multy-io/Multy-back/cmd && rm -rf multy && cd .. && make build  && cd cmd && ./$(NAME) -memprofile mem.prof && ../

setup:
	go get -u github.com/kardianos/govendor

deps:
	govendor sync

dist: TARGET_OS=linux
dist: TARGET_ARCH=amd64
dist: build

build: $(TARGETS)
	ls -lah $(TARGETS)

$(TARGETS):
	cd $(@D) && \
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build $(LD_OPTS) -o $(@F) . && \
	cd -

build-docker-images: $(DOCKER_IMAGES)

$(DOCKER_IMAGES):
	docker build --target $@ --tag $@:$(DOCKER_TAG) .

.PHONY: test
test:
	go test ./...

proto-btc-ns:
	cd ./ns-btc-protobuf && protoc --go_out=plugins=grpc:. *.proto
	cd ./ns-eth-protobuf && protoc --go_out=plugins=grpc:. *.proto

proto: proto-btc-ns

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