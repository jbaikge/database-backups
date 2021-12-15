package app

import (
	"github.com/gin-gonic/gin"
	"github.com/jbaikge/database-backups/pkg/api"
)

type Server struct {
	router          *gin.Engine
	serverService   api.ServerService
	databaseService api.DatabaseService
}

func NewServer(router *gin.Engine, serverService api.ServerService, databaseService api.DatabaseService) *Server {
	return &Server{
		router:          router,
		serverService:   serverService,
		databaseService: databaseService,
	}
}

func (s *Server) Run(listenAddress string) error {
	r := s.Routes()
	return r.Run(listenAddress)
}
