package api

type DatabaseService interface {
	Delete(int) error
	Get(int) (*Database, error)
	List() ([]Database, error)
	ListServer(int) ([]Database, error)
	New(NewDatabaseRequest) error
	Update(int, UpdateDatabaseRequest) error
}

type DatabaseRepository interface {
	CreateDatabase(NewDatabaseRequest) error
	DeleteDatabase(int) error
	GetDatabase(int) (*Database, error)
	ListDatabases() ([]Database, error)
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

func (s *databaseService) List() ([]Database, error) {
	return s.storage.ListDatabases()
}

func (s *databaseService) ListServer(serverId int) ([]Database, error) {
	all, err := s.storage.ListDatabases()
	if err != nil {
		return nil, err
	}

	dbs := make([]Database, 0, len(all))
	for _, db := range all {
		if db.ServerId == serverId {
			dbs = append(dbs, db)
		}
	}

	return dbs, nil
}

func (s *databaseService) New(database NewDatabaseRequest) error {
	return s.storage.CreateDatabase(database)
}

func (s *databaseService) Update(id int, database UpdateDatabaseRequest) error {
	return s.storage.UpdateDatabase(id, database)
}
