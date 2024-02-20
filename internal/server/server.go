package server

import (
	"net/http"

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
}

func New(addr string) *Server {
	return &Server{
		router:  echo.New(),
		address: addr,
	}
}

func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	ctx.Response().Writer.WriteHeader(statusCode)
	ctx.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTML)
	return t.Render(ctx.Request().Context(), ctx.Response().Writer)
}

func (s *Server) setupViews() {
	s.router.GET("/", func(e echo.Context) error {
		return Render(e, http.StatusOK, views.Layout())
	})
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}\n",
	}))
}

func (s *Server) setupRoutes() {
	s.router.GET("/api/:name", func(e echo.Context) error {
		name := e.Param("name")
		if name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name is required")
		}

		e.Response().WriteHeader(http.StatusOK)
		return jsoniter.NewEncoder(e.Response().Writer).Encode(map[string]string{"name": name})
	})
	s.router.GET("/api/search", func(e echo.Context) error {
		query := e.QueryParam("id")
		if query == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "id is required")
		}

		e.Response().WriteHeader(http.StatusOK)
		return jsoniter.NewEncoder(e.Response().Writer).Encode(map[string]string{"query": query})
	})
	s.router.POST("/api/register", func(e echo.Context) error {
		body := struct {
			Name string `json:"name"`
		}{}

		if err := jsoniter.NewDecoder(e.Request().Body).Decode(&body); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
		}

		if body.Name == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "name is required")
		}

		e.Response().WriteHeader(http.StatusCreated)
		return nil
	})

}

func (s *Server) Start() {
	s.setupRoutes()
	s.setupViews()
	s.setupMiddleware()

	err := make(chan error, 1)
	go func() {
		log.Info().Str("address", s.address).Msg("server started")
		err <- s.router.Start(s.address)
	}()

	if e := <-err; e != nil {
		log.Error().Err(e).Msg("server error")
	}
}
