# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . $GOPATH/src/git.loadtests.me/loadtests/loadtests/executor
ADD test_data ./test_data

# Build the outyet command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go install git.loadtests.me/loadtests/loadtests/executor

# Run the outyet command by default when the container starts.
ENTRYPOINT /go/bin/executor
