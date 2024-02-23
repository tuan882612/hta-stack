package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-h/templ"
	jsoniter "github.com/json-iterator/go"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/rs/zerolog/log"

	"hta/internal/views"
)

type Server struct {
	router  *echo.Echo
	address string
	errCh   chan error
}

func New(addr string) *Server {
	svr := &Server{
		router:  echo.New(),
		address: addr,
		errCh:   make(chan error),
	}
	svr.setupMiddleware()
	svr.setupRoutes()
	svr.setupViews()
	return svr
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (s *Server) setupViews() {
	s.router.GET("/", func(e echo.Context) error {
		return Render(e, http.StatusOK, views.Index())
	})
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "time_unix=${time_unix}, request='${method} ${uri}', status=${status}, latency=${latency_human}\n",
	}))
}

func (s *Server) setupRoutes() {
	s.router.GET("/api/:name", func(e echo.Context) error {
		name := e.Param("name")
		if name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name is required")
		}

		data := []map[string]string{
			{"name": name},
			{"name": "another " + name},
			{"name": "yet another " + name},
		}

		e.Response().WriteHeader(http.StatusOK)
		return jsoniter.NewEncoder(e.Response().Writer).Encode(data)
	})
}

func (s *Server) watch() {
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	<-sigch

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.router.Shutdown(ctx)
	if err != nil {
		s.errCh <- err
	}

	close(s.errCh)
}

func (s *Server) Start() {
	go s.watch()

	if err := s.router.Start(s.address); err != http.ErrServerClosed {
		log.Fatal().Err(err).Msg("server error")
		return
	}
	
	if err, ok := <-s.errCh; ok {
		log.Fatal().Err(err).Msg("shutdown error")
	}

	log.Info().Msg("server shutdown gracefully...")
}
