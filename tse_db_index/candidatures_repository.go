package main

type candidaturesRepository interface {
	save(votingCity *votingCity) error
}
