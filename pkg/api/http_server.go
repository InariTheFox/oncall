package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
)

type HTTPServer struct {
	Cfg        *setting.Cfg
	Listener   net.Listener
	context    context.Context
	router     *chi.Mux
	httpServer *http.Server
}

func New(cfg *setting.Cfg) (*HTTPServer, error) {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(render.SetContentType(render.ContentTypeJSON))

	s := &HTTPServer{
		Cfg:    cfg,
		router: r,
	}

	return s, nil
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.context = ctx

	s.applyRoutes()

	host := strings.TrimSuffix(strings.TrimPrefix(s.Cfg.HTTPAddr, "["), "]")
	s.httpServer = &http.Server{
		Addr:        net.JoinHostPort(host, s.Cfg.HTTPPort),
		Handler:     s.router,
		ReadTimeout: s.Cfg.ReadTimeout,
	}

	listener, err := s.getListener()
	if err != nil {
		return err
	}

	fmt.Printf("HTTP Server listen %s\n", listener.Addr().String())

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := s.httpServer.Shutdown(context.Background()); err != nil {
			fmt.Errorf("Failed to shutdown server %w\n", err)
		}
	}()

	switch s.Cfg.Protocol {
	case setting.HTTPScheme, setting.SocketScheme:
		if err := s.httpServer.Serve(listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server was shutdown gracefully")
				return nil
			}
			return err
		}
	case setting.HTTP2Scheme, setting.HTTPSScheme:
		if err := s.httpServer.ServeTLS(listener, "", ""); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server was shutdown gracefully")
				return nil
			}
			return err
		}
	default:
		panic(fmt.Sprintf("Unhandled protocol %q\n", s.Cfg.Protocol))
	}

	wg.Wait()

	return nil
}

func (s *HTTPServer) applyRoutes() {

	s.Get("/", s.Index)
}

func (s *HTTPServer) getListener() (net.Listener, error) {
	if s.Listener != nil {
		return s.Listener, nil
	}

	switch s.Cfg.Protocol {
	case setting.HTTPScheme, setting.HTTPSScheme, setting.HTTP2Scheme:
		listener, err := net.Listen("tcp", s.httpServer.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to open listener on address %s: %w\n", s.httpServer.Addr, err)
		}

		return listener, nil
	default:
		return nil, fmt.Errorf("invalid protocol %q\n", s.Cfg.Protocol)
	}
}
