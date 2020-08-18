package main

import (
	"fmt"

	"github.com/candidatos-info/descritor"
)

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

func (m *inMemoryRepository) findCandidateByEmail(email string) (*descritor.Candidatura, error) {
	for _, votingPlace := range m.db {
		for _, candidature := range votingPlace.Candidates {
			if candidature.Candidato.Email == email {
				return candidature, nil
			}
		}
	}
	return nil, fmt.Errorf("candidato com email [%s] não encontrato", email)
}
