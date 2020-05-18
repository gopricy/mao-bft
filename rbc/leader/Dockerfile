FROM golang:1.13

ADD . /go/src/visa.com/gopricy/mao-bft
RUN EXPORT GOPATH=/go && \
    cd /go/src/visa.com/gopricy/mao-bft &&\
    go build -o ./rbc/bin/leader ./rbc/leader/main.go

ENTRYPOINT ["./rbc/bin/leader"]
