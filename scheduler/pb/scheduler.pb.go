// Code generated by protoc-gen-go.
// source: pb/scheduler.proto
// DO NOT EDIT!

/*
Package pb is a generated protocol buffer package.

It is generated from these files:
	pb/scheduler.proto

It has these top-level messages:
	LoadTestReq
	LoadTestResp
	RegisterExecutorReq
	RegisterExecutorResp
*/
package pb

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

type LoadTestReq struct {
	URL                       string  `protobuf:"bytes,1,opt" json:"URL,omitempty"`
	Script                    string  `protobuf:"bytes,2,opt,name=script" json:"script,omitempty"`
	ScriptName                string  `protobuf:"bytes,3,opt,name=scriptName" json:"scriptName,omitempty"`
	RunTime                   int32   `protobuf:"varint,4,opt,name=runTime" json:"runTime,omitempty"`
	MaxWorkers                int32   `protobuf:"varint,6,opt,name=maxWorkers" json:"maxWorkers,omitempty"`
	GrowthFactor              float64 `protobuf:"fixed64,8,opt,name=growthFactor" json:"growthFactor,omitempty"`
	TimeBetweenGrowth         float64 `protobuf:"fixed64,9,opt,name=timeBetweenGrowth" json:"timeBetweenGrowth,omitempty"`
	StartingRequestsPerSecond int32   `protobuf:"varint,10,opt,name=startingRequestsPerSecond" json:"startingRequestsPerSecond,omitempty"`
	MaxRequestsPerSecond      int32   `protobuf:"varint,11,opt,name=maxRequestsPerSecond" json:"maxRequestsPerSecond,omitempty"`
}

func (m *LoadTestReq) Reset()         { *m = LoadTestReq{} }
func (m *LoadTestReq) String() string { return proto.CompactTextString(m) }
func (*LoadTestReq) ProtoMessage()    {}

type LoadTestResp struct {
	Start  *LoadTestResp_Started  `protobuf:"bytes,1,opt,name=start" json:"start,omitempty"`
	Finish *LoadTestResp_Finished `protobuf:"bytes,2,opt,name=finish" json:"finish,omitempty"`
	Error  *LoadTestResp_Errored  `protobuf:"bytes,3,opt,name=error" json:"error,omitempty"`
}

func (m *LoadTestResp) Reset()         { *m = LoadTestResp{} }
func (m *LoadTestResp) String() string { return proto.CompactTextString(m) }
func (*LoadTestResp) ProtoMessage()    {}

func (m *LoadTestResp) GetStart() *LoadTestResp_Started {
	if m != nil {
		return m.Start
	}
	return nil
}

func (m *LoadTestResp) GetFinish() *LoadTestResp_Finished {
	if m != nil {
		return m.Finish
	}
	return nil
}

func (m *LoadTestResp) GetError() *LoadTestResp_Errored {
	if m != nil {
		return m.Error
	}
	return nil
}

type LoadTestResp_Started struct {
}

func (m *LoadTestResp_Started) Reset()         { *m = LoadTestResp_Started{} }
func (m *LoadTestResp_Started) String() string { return proto.CompactTextString(m) }
func (*LoadTestResp_Started) ProtoMessage()    {}

type LoadTestResp_Finished struct {
}

func (m *LoadTestResp_Finished) Reset()         { *m = LoadTestResp_Finished{} }
func (m *LoadTestResp_Finished) String() string { return proto.CompactTextString(m) }
func (*LoadTestResp_Finished) ProtoMessage()    {}

type LoadTestResp_Errored struct {
}

func (m *LoadTestResp_Errored) Reset()         { *m = LoadTestResp_Errored{} }
func (m *LoadTestResp_Errored) String() string { return proto.CompactTextString(m) }
func (*LoadTestResp_Errored) ProtoMessage()    {}

type RegisterExecutorReq struct {
	DropletId int64 `protobuf:"varint,1,opt,name=droplet_id" json:"droplet_id,omitempty"`
	Port      int64 `protobuf:"varint,2,opt,name=port" json:"port,omitempty"`
}

func (m *RegisterExecutorReq) Reset()         { *m = RegisterExecutorReq{} }
func (m *RegisterExecutorReq) String() string { return proto.CompactTextString(m) }
func (*RegisterExecutorReq) ProtoMessage()    {}

type RegisterExecutorResp struct {
}

func (m *RegisterExecutorResp) Reset()         { *m = RegisterExecutorResp{} }
func (m *RegisterExecutorResp) String() string { return proto.CompactTextString(m) }
func (*RegisterExecutorResp) ProtoMessage()    {}

// Client API for Scheduler service

type SchedulerClient interface {
	LoadTest(ctx context.Context, in *LoadTestReq, opts ...grpc.CallOption) (Scheduler_LoadTestClient, error)
	RegisterExecutor(ctx context.Context, in *RegisterExecutorReq, opts ...grpc.CallOption) (*RegisterExecutorResp, error)
}

type schedulerClient struct {
	cc *grpc.ClientConn
}

func NewSchedulerClient(cc *grpc.ClientConn) SchedulerClient {
	return &schedulerClient{cc}
}

func (c *schedulerClient) LoadTest(ctx context.Context, in *LoadTestReq, opts ...grpc.CallOption) (Scheduler_LoadTestClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Scheduler_serviceDesc.Streams[0], c.cc, "/.Scheduler/LoadTest", opts...)
	if err != nil {
		return nil, err
	}
	x := &schedulerLoadTestClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Scheduler_LoadTestClient interface {
	Recv() (*LoadTestResp, error)
	grpc.ClientStream
}

type schedulerLoadTestClient struct {
	grpc.ClientStream
}

func (x *schedulerLoadTestClient) Recv() (*LoadTestResp, error) {
	m := new(LoadTestResp)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *schedulerClient) RegisterExecutor(ctx context.Context, in *RegisterExecutorReq, opts ...grpc.CallOption) (*RegisterExecutorResp, error) {
	out := new(RegisterExecutorResp)
	err := grpc.Invoke(ctx, "/.Scheduler/RegisterExecutor", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Scheduler service

type SchedulerServer interface {
	LoadTest(*LoadTestReq, Scheduler_LoadTestServer) error
	RegisterExecutor(context.Context, *RegisterExecutorReq) (*RegisterExecutorResp, error)
}

func RegisterSchedulerServer(s *grpc.Server, srv SchedulerServer) {
	s.RegisterService(&_Scheduler_serviceDesc, srv)
}

func _Scheduler_LoadTest_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(LoadTestReq)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(SchedulerServer).LoadTest(m, &schedulerLoadTestServer{stream})
}

type Scheduler_LoadTestServer interface {
	Send(*LoadTestResp) error
	grpc.ServerStream
}

type schedulerLoadTestServer struct {
	grpc.ServerStream
}

func (x *schedulerLoadTestServer) Send(m *LoadTestResp) error {
	return x.ServerStream.SendMsg(m)
}

func _Scheduler_RegisterExecutor_Handler(srv interface{}, ctx context.Context, codec grpc.Codec, buf []byte) (interface{}, error) {
	in := new(RegisterExecutorReq)
	if err := codec.Unmarshal(buf, in); err != nil {
		return nil, err
	}
	out, err := srv.(SchedulerServer).RegisterExecutor(ctx, in)
	if err != nil {
		return nil, err
	}
	return out, nil
}

var _Scheduler_serviceDesc = grpc.ServiceDesc{
	ServiceName: ".Scheduler",
	HandlerType: (*SchedulerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterExecutor",
			Handler:    _Scheduler_RegisterExecutor_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "LoadTest",
			Handler:       _Scheduler_LoadTest_Handler,
			ServerStreams: true,
		},
	},
}
