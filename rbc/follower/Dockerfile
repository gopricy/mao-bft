FROM golang:1.13

ADD . /go/src/visa.com/gopricy/mao-bft
RUN EXPORT GOPATH=/go && \
    cd /go/src/visa.com/gopricy/mao-bft &&\
    go build -o ./rbc/bin/follower ./rbc/follower/main.go

ENTRYPOINT ["./rbc/bin/follower"]
