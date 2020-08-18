package main

import "github.com/candidatos-info/descritor"

type candidaturesRepository interface {
	save(votingCity *votingCity) error

	findCandidateByEmail(email string) (*descritor.Candidatura, error)
}
