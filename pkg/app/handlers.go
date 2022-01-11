package app

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jbaikge/database-backups/pkg/api"
)

func (s *Server) CreateServer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var newServer api.NewServerRequest
		if err := c.BindJSON(&newServer); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		server, err := s.serverService.New(newServer)
		if err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		if server == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
		}
		c.JSON(http.StatusCreated, server)
	}
}

func (s *Server) DeleteDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := s.databaseService.Delete(id); err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) DeleteServer() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := s.serverService.Delete(id); err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) GetDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		database, err := s.databaseService.Get(id)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		if database == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		c.JSON(http.StatusOK, database)
	}
}

func (s *Server) GetServer() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		server, err := s.serverService.Get(id)
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		if server == nil {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		c.JSON(http.StatusOK, server)
	}
}

func (s *Server) ListServers() gin.HandlerFunc {
	return func(c *gin.Context) {
		servers, err := s.serverService.List()
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, servers)
	}
}

func (s *Server) Tree() gin.HandlerFunc {
	return func(c *gin.Context) {
		tree, err := s.serverService.Tree()
		if err != nil {
			c.JSON(http.StatusExpectationFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, tree)
	}
}

func (s *Server) UpdateDatabase() gin.HandlerFunc {
	return func(c *gin.Context) {
		var database api.UpdateDatabaseRequest

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := c.BindJSON(&database); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := s.databaseService.Update(id, database); err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) UpdateServer() gin.HandlerFunc {
	return func(c *gin.Context) {
		var server api.NewServerRequest

		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := c.BindJSON(&server); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := s.serverService.Update(id, server); err != nil {
			c.JSON(http.StatusPreconditionFailed, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}

func (s *Server) UpdateServerDatabases() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		if err := s.serverService.UpdateDatabases(id); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"success": true})
	}
}
