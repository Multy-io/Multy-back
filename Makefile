NAME = multy

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)
LASTTAG = $(shell git describe --tags --abbrev=0 --dirty)
GOPATH = $(shell echo "$$GOPATH")
LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.lasttag=${LASTTAG} -X main.buildtime=${BUILDTIME} -linkmode=external -w -s"

# List of all binary targets we expect from make to produce
TARGETS=cmd/multy-back/multy-back cmd/ns-btc/ns-btc cmd/ns-eth/ns-eth

TARGET_OS=
TARGET_ARCH=

all: proto build test

all-with-deps: setup deps dist

run: build
	$(foreach target,$(TARGETS), ./$(target)&) true

# Just test that built binaries can be run on this system.
canary-test: build
	$(foreach target,$(TARGETS), echo "Canary-testing $(target)" && ./$(target) --CanaryTest=true;) echo "Canary done"

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

# 'true' is to avoid failing the rule if any target file does not exist.
clean:
	$(foreach target,$(TARGETS), rm -v ./$(target);) true

#################################################################################
# Docker-related stuff goes beyond this line
#################################################################################

# List of all docker images to build and tag
DOCKER_IMAGES=multy-back multy-btc-node-service multy-eth-node-service

# The default tag, used for building images, to remove ambigulty of ':latest'
DOCKER_BUILD_TAG=$(COMMIT)
# The tag image is pushed with
DOCKER_TAG?=$(DOCKER_BUILD_TAG)
DOCKERHUB_ACCOUNT=multyio

.PHONY: docker-build-images
.PHONY: docker-retag-images
.PHONY: docker-push-images
.PHONY: docker-build-builder-image

docker-all: docker-build-images docker-retag-images docker-push-images

docker-build-images: $(DOCKER_IMAGES)

# Builds an image with tag:git_commit_hash
$(DOCKER_IMAGES):
	docker build --target $@ \
		--tag $(DOCKERHUB_ACCOUNT)/$@:$(DOCKER_BUILD_TAG) \
		--build-arg BUILD_DATE=$(BUILDTIME) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg GIT_BRANCH=$(BRANCH) \
		--build-arg GIT_TAG=$(LASTTAG) \
		.

# Explicitly set the tag: changes tag from git_commit_hash to $(DOCKER_TAG) for all images
docker-retag-images:
	$(foreach docker_image,$(DOCKER_IMAGES), docker tag $(DOCKERHUB_ACCOUNT)/$(docker_image):$(DOCKER_BUILD_TAG) $(DOCKERHUB_ACCOUNT)/$(docker_image):$(DOCKER_TAG);)

# pushes images tagged with $(DOCKER_TAG) to dockerhub
docker-push-images:
	$(foreach docker_image,$(DOCKER_IMAGES), docker push $(DOCKERHUB_ACCOUNT)/$(docker_image):$(DOCKER_TAG);)

docker-test-images: docker-build-images
	$(foreach docker_image,$(DOCKER_IMAGES), docker run $(DOCKERHUB_ACCOUNT)/$(docker_image):$(DOCKER_TAG);)

# Special case: a builder image, please note that it is not automatically retagged nor pushed to avoid 
# accidentally breaking the builds on both CI/CD and developer machine.
# Thus setting making a release tag and pushing that to docker hub must be done manually and with extra care
docker-build-builder-image:
	docker build --target multy-back-builder \
		--file Dockerfile_multy-back-builder \
		--tag $(DOCKERHUB_ACCOUNT)/multy-back-builder:$(DOCKER_BUILD_TAG) \
		--build-arg BUILD_DATE=$(BUILDTIME) \
		--build-arg GIT_COMMIT=$(COMMIT) \
		--build-arg GIT_BRANCH=$(BRANCH) \
		--build-arg GIT_TAG=$(LASTTAG) \
		.

docker-retag-builder-image:
	docker tag $(DOCKERHUB_ACCOUNT)/multy-back-builder:$(DOCKER_BUILD_TAG) $(DOCKERHUB_ACCOUNT)/multy-back-builder:$(DOCKER_TAG)

docker-push-builder-image:
	docker push $(DOCKERHUB_ACCOUNT)/multy-back-builder:$(DOCKER_TAG)