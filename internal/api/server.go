package api

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"sql-index-simulator/internal/app/handler"
	"sql-index-simulator/internal/app/repository"
)

// Server wraps HTTP server and router.
type Server struct {
	httpServer *http.Server
	engine     *gin.Engine
}

// NewServer initializes repository, handler and routes.
func NewServer() *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Static files
	r.Static("/static", "./resources")

	// Templates
	r.LoadHTMLGlob("templates/*")

	repo := repository.NewRepository()
	h := handler.NewHandler(repo)

	r.GET("/", h.GetServicesPage)
	r.GET("/hello", h.GetServicesPage)
	r.GET("/service/:id", h.GetServicePage)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	return &Server{
		httpServer: srv,
		engine:     r,
	}
}

// Start runs HTTP server.
func (s *Server) Start() error {
	logrus.Infof("HTTP server is listening on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully shuts down HTTP server.
func (s *Server) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}


