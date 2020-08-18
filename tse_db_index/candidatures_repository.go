package main

type candidaturesRepository interface {
	save(votingCity *votingCity) error

	findCandidateByEmail(email string) (*votingCity, error)
}
