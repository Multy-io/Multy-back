FROM golang:1.9.4

# RUN go get github.com/Appscrunch/Multy-back && \
#     cd $GOPATH/src/github.com/Appscrunch && \ 
#     rm -rf Multy-back && \ 
#     git clone https://github.com/Appscrunch/Multy-back.git && \ 
#     cd Multy-back && \ 
#     git pull origin master 

RUN go get github.com/Appscrunch/Multy-back && \
    rm -rf $GOPATH/src/github.com/Appscrunch/Multy-back && \
    cd $GOPATH/src/github.com/Appscrunch && \ 
    git clone https://github.com/Appscrunch/Multy-back.git && \ 
    cd Multy-back && \ 
    git checkout versions

RUN cd $GOPATH/src/github.com/Appscrunch && \
    git clone https://github.com/Appscrunch/Multy-BTC-node-service.git && \
    cd Multy-BTC-node-service  && \
    git checkout versions && \
    # git pull origin versions && \
    make all-with-deps

# RUN cd $GOPATH/src/github.com/Appscrunch/Multy-BTC-node-service && \
#     git checkout 69c5aba051237417bb110736eddf0bc56efb2639 && \
#     git pull && \
#     make all-with-deps

WORKDIR /go/src/github.com/Appscrunch/Multy-BTC-node-service/cmd

RUN echo "VERSION 02"

ENTRYPOINT $GOPATH/src/github.com/Appscrunch/Multy-BTC-node-service/cmd/client