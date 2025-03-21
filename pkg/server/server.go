package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"sync"

	"github.com/InariTheFox/oncall/pkg/api"
	"github.com/InariTheFox/oncall/pkg/services"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/InariTheFox/oncall/pkg/worker"
	"golang.org/x/sync/errgroup"
)

type Options struct {
	Listener net.Listener
}

type Server struct {
	cfg                *setting.Cfg
	childRoutines      *errgroup.Group
	context            context.Context
	shutdownFn         context.CancelFunc
	shutdownFinished   chan struct{}
	shutdownOnce       sync.Once
	isInitialized      bool
	mtx                sync.Mutex
	backgroundServices []services.BackgroundService

	HTTPServer *api.HTTPServer
}

func New(cfg *setting.Cfg, api *api.HTTPServer, worker *worker.RabbitWorker) (*Server, error) {
	s, err := newServer(cfg, api, worker)
	if err != nil {
		return nil, err
	}

	if err := s.Init(); err != nil {
		return nil, err
	}

	return s, nil
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

	for _, svc := range s.backgroundServices {
		serviceName := reflect.TypeOf(svc).String()

		s.childRoutines.Go(func() error {
			select {
			case <-s.context.Done():
				return s.context.Err()
			default:
			}

			fmt.Printf("Starting service: %s\n", serviceName)

			err := svc.Run(s.context)

			if err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("Stopped background service %s\r\nReason: %w\r\n", serviceName, err)

				return fmt.Errorf("%s run error: %w", serviceName, err)
			}

			fmt.Printf("Stopped background service %s\r\nReason: %w\r\n", serviceName, err)

			return nil
		})
	}

	return s.childRoutines.Wait()
}

func (s *Server) Shutdown(ctx context.Context, reason string) error {
	var err error

	s.shutdownOnce.Do(func() {
		fmt.Printf("Shutdown started %s\n", reason)
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

func newServer(cfg *setting.Cfg, api *api.HTTPServer, worker *worker.RabbitWorker) (*Server, error) {
	rootCtx, shutdownFn := context.WithCancel(context.Background())
	childRoutines, childCtx := errgroup.WithContext(rootCtx)

	s := &Server{
		context:          childCtx,
		childRoutines:    childRoutines,
		HTTPServer:       api,
		cfg:              cfg,
		shutdownFn:       shutdownFn,
		shutdownFinished: make(chan struct{}),
		backgroundServices: []services.BackgroundService{
			api,
			worker,
		},
	}

	return s, nil
}
