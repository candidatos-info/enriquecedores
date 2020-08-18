package main

import (
	"context"
	"fmt"

	"cloud.google.com/go/datastore"
)

type datastoreRepository struct {
	client *datastore.Client
}

func newDatastoreRepository(client *datastore.Client) candidaturesRepository {
	return &datastoreRepository{
		client: client,
	}
}

func (ds *datastoreRepository) save(votingCity *votingCity) error {
	votinLocationID := datastore.NameKey(candidaturesCollection, fmt.Sprintf("%s_%s", votingCity.State, votingCity.City), nil)
	if _, err := ds.client.Put(context.Background(), votinLocationID, &votingCity); err != nil {
		return fmt.Errorf("falha ao salvar local de votação com estado [%s] e cidade [%s], erro %q", votingCity.State, votingCity.City, err)
	}
	return nil
}
