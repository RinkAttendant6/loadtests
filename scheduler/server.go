package scheduler

import (
	"math"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/benbjohnson/clock"
	"github.com/digitalocean/godo"
	"github.com/lgpeterson/loadtests/scheduler/pb"
	"golang.org/x/net/context"
)

var _ pb.SchedulerServer = new(Server)

type Config struct {
	PullExecutorBinaryURL string
	AdvertiseListenAddr   string

	SSHKeyIDs        []godo.DropletCreateSSHKey
	DropletRegion    string
	DropletSize      string
	DropletImageSlug string

	MaxWaitExecutorOnline time.Duration

	MaxWorkerPerExecutor int
	MaxExecPSPerExecutor int

	InfluxAddr     string
	InfluxUsername string
	InfluxPassword string
	InfluxDBName   string
	InfluxSSL      bool
}

type Server struct {
	cfg *Config

	db    *DB
	clock clock.Clock
}

func NewServer(cfg *Config, db *DB) *Server {
	return &Server{cfg: cfg, db: db, clock: clock.New()}
}

func (s *Server) RegisterExecutor(ctx context.Context, req *pb.RegisterExecutorReq) (*pb.RegisterExecutorResp, error) {
	resp := &pb.RegisterExecutorResp{
		InfluxAddr:     s.cfg.InfluxAddr,
		InfluxUsername: s.cfg.InfluxUsername,
		InfluxPassword: s.cfg.InfluxPassword,
		InfluxDb:       s.cfg.InfluxDBName,
		InfluxSsl:      s.cfg.InfluxSSL,
	}
	err := s.db.RegisterExecutorUp(int(req.DropletId), int(req.Port))
	return resp, err
}

func (s *Server) LoadTest(req *pb.LoadTestReq, srv pb.Scheduler_LoadTestServer) error {
	ctx := srv.Context()
	//	ctx, cancel := context.WithCancel(srv.Context())
	//defer cancel()

	needExecutors := int(math.Ceil(
		float64(req.MaxRequestsPerSecond) / float64(s.cfg.MaxExecPSPerExecutor),
	))

	if err := s.answerPreparing(srv, needExecutors); err != nil {
		return err
	}

	executors, err := s.db.LaunchExecutors(ctx, needExecutors)
	if err != nil {
		return err
	}
	defer func() {
		logrus.Info("killing all executors")
		if err := executors.killall(); err != nil {
			logrus.WithError(err).Error("couldn't kill all executors!")
		}
	}()

	//beginCtx, timeout := context.WithTimeout(ctx, s.cfg.MaxWaitExecutorOnline)
	//defer timeout()
	err = executors.executeCommand(
		ctx,
		req.Url,
		req.Script,
		req.ScriptName,
		req.RunTime,
		req.MaxWorkers,
		req.GrowthFactor,
		req.TimeBetweenGrowth,
		req.StartingRequestsPerSecond,
		req.MaxRequestsPerSecond,
	)
	if err != nil {
		logrus.WithError(err).Error("sending command")
		s.answerErrored(srv, err)
		return nil
	}
	s.answerStarted(srv)

	completion := make(chan error, 1)
	go func() {
		defer close(completion)
		if err := executors.waitCompletion(ctx); err != nil {
			completion <- err
		}
	}()

	// maxRuntime is 125% of announced runtime
	//maxRuntime := ((time.Second * time.Duration(req.RunTime) * 125) / 100)

	select {
	case err := <-completion:

		if err != nil {
			logrus.WithError(err).Error("waiting for completeion")
			s.answerErrored(srv, err)
		} else {
			s.answerFinished(srv)
		}
		/*
			case <-ctx.Done():
			case <-time.After(maxRuntime):
				logrus.WithError(err).Error("timing out execution")
				s.answerErrored(srv, fmt.Errorf("forcing destruction of executors, max runtime elapsed: %v", maxRuntime))
		*/
	}

	return nil
}

func (s *Server) answerPreparing(srv pb.Scheduler_LoadTestServer, count int) error {
	err := srv.Send(&pb.LoadTestResp{Preparing: &pb.LoadTestResp_Preparing{Count: int32(count)}})
	if err != nil {
		logrus.WithError(err).Error("can't send message to client")
	}
	return err
}

func (s *Server) answerStarted(srv pb.Scheduler_LoadTestServer) {
	err := srv.Send(&pb.LoadTestResp{Start: &pb.LoadTestResp_Started{}})
	if err != nil {
		logrus.WithError(err).Error("can't send message to client")
	}
}

func (s *Server) answerFinished(srv pb.Scheduler_LoadTestServer) {
	err := srv.Send(&pb.LoadTestResp{Finish: &pb.LoadTestResp_Finished{}})
	if err != nil {
		logrus.WithError(err).Error("can't send message to client")
	}
}

func (s *Server) answerErrored(srv pb.Scheduler_LoadTestServer, ansErr error) {
	err := srv.Send(&pb.LoadTestResp{Error: &pb.LoadTestResp_Errored{Error: ansErr.Error()}})
	if err != nil {
		logrus.WithError(err).Error("can't send message to client")
	}
}
