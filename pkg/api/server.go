package api

import "errors"

type ServerService interface {
	Delete(int) error
	Get(int) (*Server, error)
	List() ([]Server, error)
	New(NewServerRequest) error
	Tree() ([]Tree, error)
	Update(int, NewServerRequest) error
	UpdateDatabases(int) error
}

type ServerRepository interface {
	CreateServer(NewServerRequest) error
	DeleteServer(int) error
	GetServer(int) (*Server, error)
	ListServers() ([]Server, error)
	ServerTree() ([]Tree, error)
	UpdateServer(int, NewServerRequest) error
	UpdateServerDatabases(int, []string) error
}

type serverService struct {
	storage ServerRepository
}

func NewServerService(repo ServerRepository) ServerService {
	return &serverService{
		storage: repo,
	}
}

func (s *serverService) Delete(id int) error {
	return s.storage.DeleteServer(id)
}

func (s *serverService) Get(id int) (*Server, error) {
	return s.storage.GetServer(id)
}

func (s *serverService) List() ([]Server, error) {
	return s.storage.ListServers()
}

func (s *serverService) New(server NewServerRequest) error {
	if err := s.newServerRequestValidation(server); err != nil {
		return err
	}

	return s.storage.CreateServer(server)
}

func (s *serverService) Tree() ([]Tree, error) {
	return s.storage.ServerTree()
}

func (s *serverService) Update(id int, server NewServerRequest) error {
	if err := s.newServerRequestValidation(server); err != nil {
		return err
	}

	if id == 0 {
		return errors.New("id cannot be zero")
	}

	return s.storage.UpdateServer(id, server)
}

func (s *serverService) UpdateDatabases(id int) error {
	server, err := s.storage.GetServer(id)
	if err != nil {
		return err
	}

	databases, err := server.DatabaseList()
	if err != nil {
		return err
	}

	return s.storage.UpdateServerDatabases(id, databases)
}

func (s *serverService) newServerRequestValidation(server NewServerRequest) error {
	if server.Name == "" {
		return errors.New("name is required")
	}

	if server.Host == "" {
		return errors.New("host is required")
	}

	if server.Port == 0 {
		return errors.New("port is required")
	}

	if server.Username == "" {
		return errors.New("username is required")
	}

	if server.ProxyHost != "" && server.ProxyUsername == "" {
		return errors.New("proxy_username is required")
	}

	if server.ProxyUsername != "" && server.ProxyHost == "" {
		return errors.New("proxy_host is required")
	}

	if server.ProxyHost != "" && server.ProxyIdentity == "" {
		return errors.New("proxy_identity is required")
	}

	return nil
}
