package scheduler

import (
	"github.com/digitalocean/godo"
	. "github.com/lgpeterson/loadtests/scheduler/pb"
	"golang.org/x/net/context"
)

var _ SchedulerServer = new(Server)

type Server struct {
	cloud *godo.Client

	db DB
}

func NewServer(cloud *godo.Client, db DB) *Server {

	return &Server{cloud: cloud, db: db}
}

func (s *Server) LoadTest(req *LoadTestReq, server Scheduler_LoadTestServer) error {
	ctx := server.Context()
	_ = ctx
	return nil
}

func (s *Server) RegisterExecutor(ctx context.Context, req *RegisterExecutorReq) (*RegisterExecutorResp, error) {
	resp := new(RegisterExecutorResp)
	err := s.db.SetExecutorUp(int(req.DropletId), int(req.Port))
	return resp, err
}
