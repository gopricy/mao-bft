FROM golang:1.13

COPY . /go/src/github.com/gopricy/mao-bft/

RUN export GOPATH=/go && \ 
    cd /go/src/github.com/gopricy/mao-bft/rbc/follower/server &&\
    go build -o follower

ENTRYPOINT ["/go/src/github.com/gopricy/mao-bft/rbc/follower/server/follower"]
