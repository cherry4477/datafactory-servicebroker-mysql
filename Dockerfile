FROM golang:1.5.1

COPY . /go/src/github.com/asiainfoLDP/datafactory-servicebroker-mysql

WORKDIR /go/src/github.com/asiainfoLDP/datafactory-servicebroker-mysql

RUN GO15VENDOREXPERIMENT=1 go build

EXPOSE 8001

CMD ["sh", "-c", "./datafactory-servicebroker-mysql"]
