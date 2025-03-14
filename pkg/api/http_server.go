package api

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/InariTheFox/oncall/pkg/api/routing"
	"github.com/InariTheFox/oncall/pkg/middleware"
	"github.com/InariTheFox/oncall/pkg/setting"
	"github.com/InariTheFox/oncall/pkg/web"
)

type HTTPServer struct {
	web              *web.Router
	context          context.Context
	httpSrv          *http.Server
	middlewares      []web.Handler
	namedMiddlewares []routing.RegisterNamedMiddleware

	RouteRegister routing.RouteRegister
	Cfg           *setting.Cfg
	Listener      net.Listener
}

type ServerOptions struct {
	Listener net.Listener
}

func (s *HTTPServer) AddMiddleware(middleware web.Handler) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *HTTPServer) AddNamedMiddleware(middleware routing.RegisterNamedMiddleware) {
	s.namedMiddlewares = append(s.namedMiddlewares, middleware)
}

func (s *HTTPServer) Run(ctx context.Context) error {
	s.context = ctx

	s.applyRoutes()

	host := strings.TrimSuffix(strings.TrimPrefix(s.Cfg.HTTPAddr, "["), "]")
	s.httpSrv = &http.Server{
		Addr:        net.JoinHostPort(host, s.Cfg.HTTPPort),
		Handler:     s.web,
		ReadTimeout: s.Cfg.ReadTimeout,
	}

	switch s.Cfg.Protocol {
	case setting.HTTP2Scheme, setting.HTTPSScheme:
		/*if err := s.configureTLS(); err != nil {
			return err
		}
		if s.Cfg.CertFile != "" && s.Cfg.KeyFile != "" {
			if s.Cfg.CertWatchInterval > 0 {
				s.httpSrv.TLSConfig.GetCertificate = s.GetCertificate
				go s.WatchAndUpdateCerts(ctx)
				fmt.Println("HTTP Server certificates reload feature is enabled")
			} else {
				fmt.Println("HTTP Server certificates reload feature is NOT enabled")
			}
		}*/
	default:
	}

	listener, err := s.getListener()
	if err != nil {
		return err
	}

	fmt.Printf("HTTP Server Listening on %s://%s/%s\r\n", s.Cfg.Protocol, listener.Addr().String(), s.Cfg.AppSubURL)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()

		<-ctx.Done()
		if err := s.httpSrv.Shutdown(context.Background()); err != nil {
			fmt.Printf("Failed to shutdown server\r\n%s", err)
		}
	}()

	switch s.Cfg.Protocol {
	case setting.HTTPScheme, setting.SocketScheme:
		if err := s.httpSrv.Serve(listener); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server was shutdown gracefully")
				return nil
			}
			return err
		}
	case setting.HTTP2Scheme, setting.HTTPSScheme:
		if err := s.httpSrv.ServeTLS(listener, "", ""); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				fmt.Println("server was shutdown gracefully")
				return nil
			}
			return err
		}
	default:
		panic(fmt.Sprintf("Unhandled protocol %q", s.Cfg.Protocol))
	}

	wg.Wait()

	return nil
}

func ProvideHTTPServer(
	opts ServerOptions,
	cfg *setting.Cfg,
	routeRegister routing.RouteRegister) (*HTTPServer, error) {
	w := web.NewRouter()

	server := &HTTPServer{
		Cfg:           cfg,
		Listener:      opts.Listener,
		RouteRegister: routeRegister,
		web:           w,
	}

	return server, nil
}

func (s *HTTPServer) addMiddlewaresAndStaticRoutes() {
	m := s.web

	for _, mw := range s.middlewares {
		m.Use(mw)
	}
}

func (s *HTTPServer) applyRoutes() {
	// start with middlewares & static routes
	s.addMiddlewaresAndStaticRoutes()
	// then add view routes & api routes
	s.RouteRegister.Register(s.web, s.namedMiddlewares...)
	// lastly not found route
	s.web.NotFound(middleware.ProvideRouteOperationName("notfound"), s.NotFoundHandler)
}

func (s *HTTPServer) getListener() (net.Listener, error) {
	if s.Listener != nil {
		return s.Listener, nil
	}

	switch s.Cfg.Protocol {
	case setting.HTTPScheme, setting.HTTPSScheme, setting.HTTP2Scheme:
		listener, err := net.Listen("tcp", s.httpSrv.Addr)
		if err != nil {
			return nil, fmt.Errorf("failed to open listener on address %s: %w", s.httpSrv.Addr, err)
		}
		return listener, nil
	case setting.SocketScheme:
		listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: s.Cfg.SocketPath, Net: "unix"})
		if err != nil {
			return nil, fmt.Errorf("failed to open listener for socket %s: %w", s.Cfg.SocketPath, err)
		}

		// Make socket writable by group
		// nolint:gosec
		if err := os.Chmod(s.Cfg.SocketPath, os.FileMode(s.Cfg.SocketMode)); err != nil {
			return nil, fmt.Errorf("failed to change socket mode %d: %w", s.Cfg.SocketMode, err)
		}

		// golang.org/pkg/os does not have chgrp
		// Changing the gid of a file without privileges requires that the target group is in the group of the process and that the process is the file owner
		if err := os.Chown(s.Cfg.SocketPath, -1, s.Cfg.SocketGid); err != nil {
			return nil, fmt.Errorf("failed to change socket group id %d: %w", s.Cfg.SocketGid, err)
		}

		return listener, nil
	default:
		fmt.Printf("Invalid protocol %s", s.Cfg.Protocol)
		return nil, fmt.Errorf("invalid protocol %q", s.Cfg.Protocol)
	}
}
