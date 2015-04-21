# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

EXPOSE 50051
ENTRYPOINT ["/go/bin/executord", "50051"]

RUN go get github.com/tools/godep 

ADD . $GOPATH/src/github.com/lgpeterson/loadtests/
WORKDIR $GOPATH/src/github.com/lgpeterson/loadtests/executor
RUN godep go build -o $GOPATH/bin/executord github.com/lgpeterson/loadtests/executor

