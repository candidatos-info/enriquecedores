package main

type candidaturesRepository interface {
	save(votingCity *votingCity) error

	findVotingCityByCandidateEmail(email string) (*votingCity, error)
}
