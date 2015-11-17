# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

EXPOSE 50051
ENTRYPOINT ["/go/bin/schedulerd"]

RUN go get github.com/tools/godep

ADD . $GOPATH/src/github.com/lgpeterson/loadtests/
WORKDIR $GOPATH/src/github.com/lgpeterson/loadtests/
RUN godep go build -o $GOPATH/bin/schedulerd github.com/lgpeterson/loadtests/cmd/schedulerd

