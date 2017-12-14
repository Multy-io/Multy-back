FROM golang:1.9.2

RUN mkdir -p go/src/github.com/Appscrunch/Multy-back
WORKDIR /go/src/github.com/Appscrunch/Multy-back

COPY . /go/src/github.com/Appscrunch/Multy-back

# TODO: add ssl creds
#COPY ssl ssl

EXPOSE 7778

CMD ["make", "all"]