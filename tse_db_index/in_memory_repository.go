package main

type inMemoryRepository struct {
	db map[string]*votingCity
}

func newInMemoryRepository() candidaturesRepository {
	return &inMemoryRepository{
		db: make(map[string]*votingCity),
	}
}

func (m *inMemoryRepository) save(votingCity *votingCity, id string) error {
	m.db[id] = votingCity
	return nil
}
