FROM golang:1.5.1

COPY . /go/src/github.com/asiainfoLDP/datafactory-servicebroker-mysql

WORKDIR /go/src/github.com/asiainfoLDP/datafactory-servicebroker-mysql

RUN go get github.com/tools/godep

RUN godep go build
    
EXPOSE 8001

CMD ["sh", "-c", "./datafactory-servicebroker-mysql"]
