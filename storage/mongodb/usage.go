package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type usageRepo struct {
	col *mongo.Collection
}

type Usage struct {
	ID        primitive.ObjectID `bson:"_id"`
	IPAddress string             `bson:"ip_address"`
	Usage     int                `bson:"usage"`
}

type UsageStorageI interface {
	GetUsage(ctx context.Context, ipAddr string) (*Usage, error)
	CreateUsage(ctx context.Context, usage *Usage) (string, error)
}

func NewUsage(db *mongo.Database) UsageStorageI {
	return &usageRepo{
		col: db.Collection("usage"),
	}
}

func (u *usageRepo) GetUsage(ctx context.Context, ipAddr string) (*Usage, error) {
	var res Usage

	err := u.col.FindOne(ctx, bson.M{"ip_address": ipAddr}).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &res, nil
}

func (u *usageRepo) CreateUsage(ctx context.Context, usage *Usage) (string, error) {

	use, err := u.GetUsage(context.Background(), usage.IPAddress)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", err
	}

	if use == nil {
		usage.ID = primitive.NewObjectID()
		res, err := u.col.InsertOne(ctx, usage)
		return res.InsertedID.(primitive.ObjectID).Hex(), err
	}

	_, err = u.col.UpdateOne(ctx, bson.M{"ip_address": usage.IPAddress}, bson.M{"$inc": bson.M{"usage": 1}})
	return "", err
}
