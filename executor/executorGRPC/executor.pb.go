// Code generated by protoc-gen-go.
// source: executor.proto
// DO NOT EDIT!

/*
Package executorGRPC is a generated protocol buffer package.

It is generated from these files:
	executor.proto

It has these top-level messages:
	StatusMessage
	CommandMessage
*/
package executorGRPC

import proto "github.com/golang/protobuf/proto"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal

type StatusMessage struct {
	Status string `protobuf:"bytes,1,opt,name=status" json:"status,omitempty"`
	Error  string `protobuf:"bytes,2,opt,name=error" json:"error,omitempty"`
}

func (m *StatusMessage) Reset()         { *m = StatusMessage{} }
func (m *StatusMessage) String() string { return proto.CompactTextString(m) }
func (*StatusMessage) ProtoMessage()    {}

type CommandMessage struct {
	IP         string `protobuf:"bytes,1,opt" json:"IP,omitempty"`
	Script     string `protobuf:"bytes,2,opt,name=script" json:"script,omitempty"`
	ScriptName string `protobuf:"bytes,3,opt,name=scriptName" json:"scriptName,omitempty"`
}

func (m *CommandMessage) Reset()         { *m = CommandMessage{} }
func (m *CommandMessage) String() string { return proto.CompactTextString(m) }
func (*CommandMessage) ProtoMessage()    {}

func init() {
}

// Client API for Commander service

type CommanderClient interface {
	ExecuteCommand(ctx context.Context, in *CommandMessage, opts ...grpc.CallOption) (*StatusMessage, error)
}

type commanderClient struct {
	cc *grpc.ClientConn
}

func NewCommanderClient(cc *grpc.ClientConn) CommanderClient {
	return &commanderClient{cc}
}

func (c *commanderClient) ExecuteCommand(ctx context.Context, in *CommandMessage, opts ...grpc.CallOption) (*StatusMessage, error) {
	out := new(StatusMessage)
	err := grpc.Invoke(ctx, "/executorGRPC.Commander/ExecuteCommand", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Commander service

type CommanderServer interface {
	ExecuteCommand(context.Context, *CommandMessage) (*StatusMessage, error)
}

func RegisterCommanderServer(s *grpc.Server, srv CommanderServer) {
	s.RegisterService(&_Commander_serviceDesc, srv)
}

func _Commander_ExecuteCommand_Handler(srv interface{}, ctx context.Context, buf []byte) (proto.Message, error) {
	in := new(CommandMessage)
	if err := proto.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(CommanderServer).ExecuteCommand(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Commander_serviceDesc = grpc.ServiceDesc{
	ServiceName: "executorGRPC.Commander",
	HandlerType: (*CommanderServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ExecuteCommand",
			Handler:    _Commander_ExecuteCommand_Handler,
		},
	},
	Streams: []grpc.StreamDesc{},
}
