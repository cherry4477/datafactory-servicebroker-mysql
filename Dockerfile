FROM golang:1.5.1

COPY . /go/src/github.com/asiainfoLDP/broker_mysql

WORKDIR /go/src/github.com/asiainfoLDP/broker_mysql

RUN go get github.com/tools/godep \
    && $GOPATH/bin/godep restore \
    && go build
    
EXPOSE 8001

CMD ["sh", "-c", "./broker_mysql"]
