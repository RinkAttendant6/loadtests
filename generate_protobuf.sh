#!/bin/sh

go build github.com/golang/protobuf/protoc-gen-go

PATH=$PATH:. protoc -I. --go_out=plugins=grpc:scheduler/ pb/scheduler.proto
PATH=$PATH:. protoc -I. --go_out=plugins=grpc:executor/ pb/executor.proto
rm protoc-gen-go
