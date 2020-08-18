package main

import "fmt"

type inMemoryRepository struct {
	db map[string]*votingCity
}

func newInMemoryRepository() candidaturesRepository {
	return &inMemoryRepository{
		db: make(map[string]*votingCity),
	}
}

func (m *inMemoryRepository) save(votingCity *votingCity) error {
	id := fmt.Sprintf("%s_%s", votingCity.State, votingCity.City)
	m.db[id] = votingCity
	return nil
}
