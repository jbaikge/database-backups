package api

type DatabaseService interface {
	Delete(int) error
	Get(int) (*Database, error)
	List(int) ([]Database, error)
	New(NewDatabaseRequest) error
	Update(int, UpdateDatabaseRequest) error
}

type DatabaseRepository interface {
	CreateDatabase(NewDatabaseRequest) error
	DeleteDatabase(int) error
	GetDatabase(int) (*Database, error)
	ListDatabases(int) ([]Database, error)
	UpdateDatabase(int, UpdateDatabaseRequest) error
}

type databaseService struct {
	storage DatabaseRepository
}

func NewDatabaseService(repo DatabaseRepository) DatabaseService {
	return &databaseService{
		storage: repo,
	}
}

func (s *databaseService) Delete(id int) error {
	return s.storage.DeleteDatabase(id)
}

func (s *databaseService) Get(id int) (*Database, error) {
	return s.storage.GetDatabase(id)
}

func (s *databaseService) List(serverId int) ([]Database, error) {
	return s.storage.ListDatabases(serverId)
}

func (s *databaseService) New(database NewDatabaseRequest) error {
	return s.storage.CreateDatabase(database)
}

func (s *databaseService) Update(id int, database UpdateDatabaseRequest) error {
	return s.storage.UpdateDatabase(id, database)
}
