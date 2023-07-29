package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewClient established connection to a mongodb instance using provideid URI and auth credentials.
func NewClient(url, username, password string) (*mongo.Database, error) {
	opts := options.Client().ApplyURI(url)
	if username != "" && password != "" {
		opts.SetAuth(options.Credential{
			Username: username,
			Password: password,
		})
	}
	opts.SetMaxPoolSize(10)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, err
	}

	// Ping the MongoDB server to ensure the connection is established
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database("zohiddev")

	return db, nil
}
