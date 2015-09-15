#!/bin/sh

protoc --go_out=plugins=grpc:scheduler/ pb/scheduler.proto
protoc --go_out=plugins=grpc:executor/ pb/executor.proto
