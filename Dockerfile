FROM golang:1.5.1

WORKDIR /go/src/broker_mysql
ADD . /go/src/broker_mysql/

RUN go get github.com/tools/godep
RUN godep restore
RUN godep go install

EXPOSE 8100

ENV SERVICE_NAME=broker_mysql

ENTRYPOINT ["/go/bin/broker_mysql"]
