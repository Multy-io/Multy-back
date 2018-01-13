FROM golang:1.9.2

RUN mkdir -p go/src/github.com/Appscrunch/Multy-back
WORKDIR /go/src/github.com/Appscrunch/Multy-back

COPY . /go/src/github.com/Appscrunch/Multy-back

RUN touch cmd/rpc.cert && \
    touch cmd/multy.config

# TODO: add ssl creds
#COPY ssl ssl

CMD ["make", "all"]