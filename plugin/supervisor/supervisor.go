package supervisor

import (
	"context"
	"fmt"
	"github.com/cenkalti/backoff"
	"github.com/mesanine/gaffer/config"
	"github.com/mesanine/gaffer/event"
	"github.com/mesanine/gaffer/log"
	"github.com/mesanine/gaffer/service"

	"github.com/mesanine/gaffer/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

const (
	BackoffInterval = 1000 * time.Millisecond
	StatsInterval   = 2000 * time.Millisecond
)

// Supervisor implements a lightweight daemon for controlling
// containers with the runc executable.
type Supervisor struct {
	runcs  map[string]*Runc
	cancel map[string]context.CancelFunc
	db     *store.FSStore
	config config.Config
	err    chan error
	stop   chan bool
}

// New creates a new supervisor
func New() *Supervisor {
	return &Supervisor{
		runcs:  map[string]*Runc{},
		cancel: map[string]context.CancelFunc{},
		err:    make(chan error),
		db:     nil,
	}
}

func (s *Supervisor) Name() string { return "supervisor" }

func (s *Supervisor) Configure(cfg config.Config) error {
	s.db = store.New(cfg, "services")
	services, err := s.db.Services()
	if err != nil {
		return err
	}
	for _, svc := range services {
		s.runcs[svc.Id] = NewRunc(svc.Id, svc.Bundle, cfg.RuncRoot)
	}
	s.config = cfg
	return nil
}

func (s *Supervisor) RPC() *grpc.ServiceDesc { return &_RPC_serviceDesc }

func (s *Supervisor) Run(eb *event.EventBus) error {
	// Launch all registered containers
	s.init(eb)
	ticker := time.NewTicker(900 * time.Millisecond)
	var stopped int
	for stopped < len(s.runcs) {
		select {
		case err := <-s.err:
			if err != nil {
				log.Log.Error("supervisor runtime error", zap.Error(err))
				return err
			}
			stopped++
		case <-ticker.C:
			// periodically publish container metrics
			// via the eventbus
			for name, runc := range s.runcs {
				stats, err := runc.Stats()
				if err != nil {
					log.Log.Warn(fmt.Sprintf("could not collect stats for container %s: %s", name, err))
					continue
				}
				eb.Push(event.New(
					event.SERVICE_METRICS,
					event.WithID(name),
					event.WithStats(*stats),
				))
			}
		}
	}
	return nil
}

func (s *Supervisor) Stop() error {
	for name, cancelFn := range s.cancel {
		log.Log.Warn(fmt.Sprintf("canceling runc service %s", name))
		// Cancel each runc backoff context
		// causing each container to not be
		// restarted when killed.
		cancelFn()
		if err := s.runcs[name].Stop(); err != nil {
			// If we can't stop a container we will log it but continue
			// trying since the entire process is being shutdown.
			log.Log.Error(fmt.Sprintf("failed to cancel service %s: %s", name, err.Error()))
		} else {
			log.Log.Warn(fmt.Sprintf("killed service %s", name))
		}
	}
	// Signial stop to the Run() function
	return nil
}

func (s *Supervisor) Status(ctx context.Context, req *StatusRequest) (*StatusResponse, error) {
	resp := &StatusResponse{
		Services: []*service.Service{},
	}
	services, err := s.db.Services()
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		stats, err := s.runcs[svc.Id].Stats()
		if err != nil {
			return nil, err
		}
		svc = service.WithStats(*stats)(svc)
		resp.Services = append(resp.Services, &svc)
	}
	return resp, nil
}

func (s *Supervisor) Restart(ctx context.Context, req *RestartRequest) (*RestartResponse, error) {
	rc, ok := s.runcs[req.Id]
	if !ok {
		return nil, fmt.Errorf("no container with id %s exists", req.Id)
	}
	// Kill the underlying runc app
	// causing the supervisor to start it again.
	err := rc.Stop()
	if err != nil {
		return nil, err
	}
	return &RestartResponse{}, nil
}

func (s *Supervisor) init(eb *event.EventBus) {
	for name, rc := range s.runcs {
		if _, ok := s.cancel[name]; ok {
			panic(fmt.Sprintf("container %s was already registered", name))
		}
		ctx, cancelFn := context.WithCancel(context.Background())
		s.cancel[name] = cancelFn
		go func(ctx context.Context, rc *Runc, name string) {
			s.err <- backoff.RetryNotify(
				func() error {
					log.Log.Info(fmt.Sprintf("launching runc container %s", name))
					eb.Push(
						event.New(
							event.SERVICE_STARTED,
							event.WithID(name),
						),
					)
					code, err := rc.Run()
					var msg string
					if err != nil {
						msg = err.Error()
					}
					return fmt.Errorf("container %s exited with code %d: %s", name, code, msg)
				},
				backoff.WithContext(backoff.NewConstantBackOff(BackoffInterval), ctx),
				func(err error, d time.Duration) {
					eb.Push(event.New(
						event.SERVICE_EXITED,
						event.WithID(name),
					))
					log.Log.Warn(err.Error(), zap.Duration("runtime", d))
				},
			)
		}(ctx, rc, name)
	}
}
