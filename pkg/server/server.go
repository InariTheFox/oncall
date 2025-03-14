package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"

	"golang.org/x/sync/errgroup"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/registry"
	"github.com/InariTheFox/oncall/pkg/setting"
)

type Options struct {
	HomePath    string
	PidFile     string
	Version     string
	Commit      string
	BuildBranch string
	Listener    net.Listener
}

func New(
	opts Options,
	cfg *setting.Cfg,
	httpServer *api.HTTPServer,
	backgroundServiceProvider registry.BackgroundServiceRegistry) (*Server, error) {
	s, err := newServer(opts, cfg, httpServer, backgroundServiceProvider)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
}

type Server struct {
	context          context.Context
	shutdownFn       context.CancelFunc
	childRoutines    *errgroup.Group
	cfg              *setting.Cfg
	shutdownOnce     sync.Once
	shutdownFinished chan struct{}
	isInitialized    bool
	mtx              sync.Mutex

	version            string
	commit             string
	buildBranch        string
	backgroundServices []registry.BackgroundService

	HTTPServer *api.HTTPServer
}

func (s *Server) Init() error {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	if s.isInitialized {
		return nil
	}

	s.isInitialized = true

	return nil
}

func (s *Server) Run() error {
	defer close(s.shutdownFinished)

	if err := s.Init(); err != nil {
		return err
	}

	services := s.backgroundServices

	for _, svc := range services {
		service := svc
		serviceName := reflect.TypeOf(service).String()
		s.childRoutines.Go(func() error {
			select {
			case <-s.context.Done():
				return s.context.Err()
			default:
			}

			fmt.Printf("Starting background service %s\r\n", serviceName)
			err := service.Run(s.context)

			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("Stopped background service %s\r\nReason: %s\r\n", serviceName, err)
				return fmt.Errorf("%s run error: %w", serviceName, err)
			}

			fmt.Printf("Stopped background service %s\r\nReason: %s\r\n", serviceName, err)
			return nil
		})
	}

	fmt.Println("Waiting on services...")
	return s.childRoutines.Wait()
}

// Shutdown initiates a graceful shutdown. This shuts down all
// running background services. Since Run blocks Shutdown supposed to
// be run from a separate goroutine.
func (s *Server) Shutdown(ctx context.Context, reason string) error {
	var err error
	s.shutdownOnce.Do(func() {
		fmt.Println("Shutdown started")
		// Call cancel func to stop background services.
		s.shutdownFn()
		// Wait for server to shut down
		select {
		case <-s.shutdownFinished:
			fmt.Println("Finished waiting for server to shut down")
		case <-ctx.Done():
			fmt.Println("Timed out while waiting for server to shut down")
			err = fmt.Errorf("timeout waiting for shutdown")
		}
	})

	return err
}

func newServer(
	opts Options,
	cfg *setting.Cfg,
	httpServer *api.HTTPServer,
	backgroundServiceProvider registry.BackgroundServiceRegistry) (*Server, error) {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	s := &Server{
		context:            childCtx,
		childRoutines:      childRoutines,
		HTTPServer:         httpServer,
		shutdownFn:         shutdownFn,
		shutdownFinished:   make(chan struct{}),
		cfg:                cfg,
		version:            opts.Version,
		commit:             opts.Commit,
		buildBranch:        opts.BuildBranch,
		backgroundServices: backgroundServiceProvider.GetServices(),
	}

	return s, nil
}
