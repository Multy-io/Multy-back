FROM golang:1.9.4

RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc && go get firebase.google.com/go && go get google.golang.org/api/option

RUN go get github.com/Appscrunch/Multy-back
RUN go get github.com/satori/go.uuid

RUN cd $GOPATH/src/github.com/Appscrunch && \ 
    rm -rf Multy-back && \ 
    git clone https://github.com/Appscrunch/Multy-back.git && \ 
    cd Multy-back && \
    git checkout release_1.1 && \
    git pull origin release_1.1


RUN cd $GOPATH/src/github.com/Appscrunch/Multy-back && \ 
    make all-with-deps 
# make all-docker 

WORKDIR /go/src/github.com/Appscrunch/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-back/cmd/multy
