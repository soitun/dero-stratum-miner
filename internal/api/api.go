package api

import (
	"context"
	"fmt"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/go-logr/logr"
	"github.com/stratumfarm/dero-stratum-miner/internal/config"
	miner "github.com/stratumfarm/dero-stratum-miner/internal/dero-stratum-miner"
	"github.com/stratumfarm/dero-stratum-miner/internal/version"
	"go.neonxp.dev/jsonrpc2/rpc"
	"go.neonxp.dev/jsonrpc2/transport"
)

type Server struct {
	ctx       context.Context
	cancel    context.CancelFunc
	listen    string
	startTime time.Time
	r         *rpc.RpcServer
	m         *miner.Client
}

func New(ctx context.Context, m *miner.Client, cfg *config.API, logr logr.Logger) *Server {
	ctx, cancel := context.WithCancel(ctx)
	r := rpc.New(
		rpc.WithLogger(&logger{logr}),
		rpc.WithTransport(&transport.HTTP{Bind: cfg.Listen, CORSOrigin: "*", Parallel: true}), // TODO: change to tcp?
	)
	s := &Server{
		ctx:    ctx,
		cancel: cancel,
		listen: cfg.Listen,
		r:      r,
		m:      m,
	}
	s.r.Register("miner_getstat1", rpc.H(s.MinerStats))
	return s
}

func (s *Server) Serve() error {
	s.startTime = time.Now()
	return s.r.Run(s.ctx)
}

func (s *Server) Close() error {
	if s.cancel != nil {
		s.cancel()
	}
	return nil
}

func (s *Server) MinerStats(ctx context.Context, args *any) (MinerStatRes, error) {
	m := MinerStat{
		Version:  version.Version,
		Runtime:  int(time.Since(s.startTime).Seconds()),
		Accepted: s.m.GetAcceptedShares(),
		Rejected: s.m.GetRejectedShares(),
		Hashrates: []string{
			fmt.Sprintf("%d", s.m.GetHashrate()),
		},
		Pool: s.m.GetPoolURL(),
	}
	return m.Res(), nil
}

type MinerStatRes []string

type MinerStat struct {
	Version   string   // version string
	Runtime   int      // runtime in seconds, can be 0
	Accepted  uint64   // accepted shares
	Rejected  uint64   // rejected shares
	Hashrates []string // hashrate in hashes
	Pool      string   // pool url
}

func (m *MinerStat) Res() MinerStatRes {
	return []string{
		fmt.Sprintf("%s %s", path.Base(os.Args[0]), m.Version),
		strconv.Itoa(m.Runtime),
		fmt.Sprintf("%d;%d;0", m.Accepted, m.Rejected),
		strings.Join(m.Hashrates, ";"),
		"0",
		"off",
		"0;0",
		m.Pool,
		"0;0;0;0",
	}
}
