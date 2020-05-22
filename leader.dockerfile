FROM golang:1.13

COPY . /go/src/github.com/gopricy/mao-bft/

RUN export GOPATH=/go && \ 
    cd /go/src/github.com/gopricy/mao-bft/rbc/leader/server &&\
    go build -o leader

ENTRYPOINT ["/go/src/github.com/gopricy/mao-bft/rbc/leader/server/leader"]
