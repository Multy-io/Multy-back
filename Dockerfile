FROM golang:1.9.4

RUN go get golang.org/x/net/context
RUN go get google.golang.org/grpc

RUN go get github.com/Appscrunch/Multy-back


RUN cd $GOPATH/src/github.com/Appscrunch && \ 
    rm -rf Multy-back && \ 
    git clone https://github.com/Appscrunch/Multy-back.git && \ 
    cd Multy-back && \
    git checkout versions && \
    git pull origin
# git pull origin master

RUN go get -d github.com/Appscrunch/Multy-back-exchange-service 


RUN cd $GOPATH/src/github.com/Appscrunch/Multy-back && \ 
    make all-with-deps && \ 
    make all-docker 

WORKDIR /go/src/github.com/Appscrunch/Multy-back/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-back/cmd/multy
