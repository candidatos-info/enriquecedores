package main

import (
	"context"
	"fmt"
	"time"

	"github.com/candidatos-info/descritor"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gopkg.in/mgo.v2/bson"
)

const (
	timeout = 10 // in seconds
)

//Client manages all iteractions with mongodb
type Client struct {
	client *mongo.Client
	dbName string
}

//New returns an db connection instance that can be used for CRUD opetations
func New(dbURL, dbName string) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(dbURL))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB at link [%s], error %v", dbURL, err)
	}
	return &Client{
		client: client,
		dbName: dbName,
	}, nil
}

// UpdateCandidate sets candidate's recurrent flag
func (c *Client) UpdateCandidate(year int, legalCode string, recurrent bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	filter := bson.M{"legal_code": legalCode, "year": year}
	update := bson.M{"$set": bson.M{"recurrent": recurrent}}
	if _, err := c.client.Database(c.dbName).Collection(descritor.CandidaturesCollection).UpdateOne(ctx, filter, update); err != nil {
		return fmt.Errorf("falha ao setar flag de recorrência no candidato [%s], error %v", legalCode, err)
	}
	return nil
}
