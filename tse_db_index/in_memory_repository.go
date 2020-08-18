package main

import (
	"fmt"
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

func (m *inMemoryRepository) findVotingCityByCandidateEmail(email string) (*votingCity, error) {
	for _, votingPlace := range m.db {
		for _, candidature := range votingPlace.Candidates {
			if candidature.Candidato.Email == email {
				return votingPlace, nil
			}
		}
	}
	return nil, fmt.Errorf("candidato com email [%s] não encontrato", email)
}
