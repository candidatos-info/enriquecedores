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

func (ds *datastoreRepository) save(votingCity *votingCity, id string) error {
	votinLocationID := datastore.NameKey(candidaturesCollection, id, nil)
	if _, err := ds.client.Put(context.Background(), votinLocationID, &votingCity); err != nil {
		return fmt.Errorf("falha ao salvar local de votação com id [%s], erro %q", id, err)
	}
	return nil
}
