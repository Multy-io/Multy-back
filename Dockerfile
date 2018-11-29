# Builder image that builds all the multy-back and all node services
# multyio/multy-back-builder has all dependencies cached
FROM multyio/multy-back-builder as builder

WORKDIR $GOPATH/src/github.com/Multy-io/Multy-back
# Build an image from sources of local directory
COPY . $GOPATH/src/github.com/Multy-io/Multy-back
RUN make build

# Base image for all images that run application
FROM alpine:3.8 as base

FROM base as multy-back
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/multy-back/multy-back /multy-back
ENTRYPOINT /multy-back

FROM base as multy-btc-node-service
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/ns-btc/ns-btc /ns-btc
ENTRYPOINT /ns-btc

FROM base as multy-eth-node-service
COPY --from=builder /go/src/github.com/Multy-io/Multy-back/cmd/ns-eth/ns-eth /ns-eth
ENTRYPOINT /ns-eth