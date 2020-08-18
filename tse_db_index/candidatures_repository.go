package main

type candidaturesRepository interface {
	save(votingCity *votingCity, id string) error
}
