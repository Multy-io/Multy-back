NAME = multy

BRANCH = $(shell git rev-parse --abbrev-ref HEAD)
COMMIT = $(shell git rev-parse --short HEAD)
BUILDTIME = $(shell date +%Y-%m-%dT%T%z)

LD_OPTS = -ldflags="-X main.branch=${BRANCH} -X main.commit=${COMMIT} -X main.buildtime=${BUILDTIME} -w"

all: setup deps build

run:
	./cmd/$(NAME)

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
	GOOS=linux $(BUILD_CMD) go build $(LD_OPTS) -o ./dist/$(NAME) .