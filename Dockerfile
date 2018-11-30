# Builder image that builds all the multy-back and all node services
# multyio/multy-back-builder has all dependencies cached
FROM multyio/multy-back-builder as builder

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
# Build an image from sources of local directory
COPY . $GOPATH/src/github.com/Multy-io/Multy-back
RUN go get -v -d ./...
RUN make build

# Base image for all images with executable application
FROM alpine:3.8 as base
# Common stuff to put into all derived containers
ONBUILD ARG BUILD_DATE
ONBUILD ARG GIT_COMMIT
ONBUILD ARG GIT_BRANCH
ONBUILD ARG GIT_TAG
ONBUILD LABEL org.label-schema.build-date="$BUILD_DATE"
ONBUILD LABEL org.label-schema.git-branch="$GIT_BRANCH"
ONBUILD LABEL org.label-schema.git-commit="$GIT_COMMIT"
ONBUILD LABEL org.label-schema.git-tag="$GIT_TAG"


FROM base as multy-back
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/multy-back/multy-back /multy-back
ENTRYPOINT /multy-back


FROM base as multy-btc-node-service
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/ns-btc/ns-btc /ns-btc
ENTRYPOINT /ns-btc


FROM base as multy-eth-node-service
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/ns-eth/ns-eth /ns-eth
ENTRYPOINT /ns-eth