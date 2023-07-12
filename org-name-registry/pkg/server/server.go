package server

import (
	"fmt"
	"github.com/Chamindu36/organization-name-registry-service/org-name-registry/pkg/clients"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"net/http"
)

type Options struct {
	EnableAuth       bool
	Port             int
	MiddlewareConfig MiddlewareConfig
}

type RouterFunc func(*mux.Router)

type Server struct {
	opt    *Options
	logger *zap.SugaredLogger
	router *mux.Router
}

type HandlerFunc func(w http.ResponseWriter, req *http.Request)

var (
	muxDispatcher = mux.NewRouter()
)

func NewMuxRouter(client *clients.AuthClient, routes Routes, opt Options, logger *zap.SugaredLogger) *Server {

	//Adding ResponseHeaderMW to server
	muxDispatcher.Use(ResponseHeadersMiddleware(map[string]string{
		"Content-Type": "application/json",
	}))
	//Adding OIDC Middleware
	muxDispatcher.Use(OidcMiddleware(client.AuthenticationClient()))

	// Mapping of the routes to the mux server
	addRouteFunc := func(r *mux.Router, route Route) {
		newRoute := r.NewRoute()
		if len(route.Path) > 0 {
			newRoute.Path(route.Path)
		} else if len(route.PathPrefix) > 0 {
			newRoute.PathPrefix(route.PathPrefix)
		}
		if len(route.Queries) > 0 {
			var pairs []string
			for _, q := range route.Queries {
				pairs = append(pairs, q.Name)
				pairs = append(pairs, q.Pattern)
			}
			newRoute.Queries(pairs...)
		}
		if len(route.Methods) > 0 {
			newRoute.Methods(route.Methods...)
		}
		newRoute.HandlerFunc(route.HandlerFunc)
	}
	for _, route := range routes {

		addRouteFunc(muxDispatcher, route)

	}
	//ADd logger
	logger = logger.Named("http-server")

	srv := &Server{
		opt:    &opt,
		logger: logger,
		router: muxDispatcher,
	}

	return srv
}

func (s *Server) Serve(opt Options) {
	addr := fmt.Sprintf(":%d", s.opt.Port)
	srv := http.Server{
		Addr:    addr,
		Handler: TraceMiddleware()(LoggingMiddleware(s.opt.MiddlewareConfig.LoggingMiddlewareConfig, s.logger)(s.router)),
	}

	s.logger.Infof("Serving on %s", addr)
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		s.logger.Errorf("ListenAndServe error: %v", err)
	}
}
