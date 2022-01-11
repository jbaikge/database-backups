package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (s *Server) Routes() *gin.Engine {
	router := s.router

	v1 := router.Group("/v1")
	{
		v1.GET("/ping", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "pong"})
		})
		v1.GET("/tree", s.Tree())
		databases := v1.Group("/databases")
		{
			databases.GET("/:id", s.GetDatabase())
			databases.PUT("/:id", s.UpdateDatabase())
			databases.DELETE("/:id", s.DeleteDatabase())
		}
		servers := v1.Group("/servers")
		{
			servers.GET("", s.ListServers())
			servers.POST("", s.CreateServer())
			servers.GET("/:id", s.GetServer())
			servers.PUT("/:id", s.UpdateServer())
			servers.DELETE("/:id", s.DeleteServer())
		}
	}

	return router
}
