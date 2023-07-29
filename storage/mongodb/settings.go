package mongodb

import (
	"context"
	"errors"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Keys struct {
	ID        primitive.ObjectID `bson:"_id"`
	Name      string             `bson:"name"`
	SSHHash   string             `bson:"ssh_hash"`
	CreatedAt string             `bson:"created_at"`
}

func (u *userRepo) SetSubdomainAndSShKey(c context.Context, id, subdomain string, key *Keys) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	_, err = u.col.UpdateOne(c, bson.M{"_id": ID}, bson.M{"$set": bson.M{"subdomain": subdomain}, "$push": bson.M{"keys": bson.M{"_id": key.ID, "ssh_hash": key.SSHHash, "name": key.Name, "created_at": key.CreatedAt}}})

	return err
}

func (u *userRepo) PushNewKey(c context.Context, id string, key *Keys) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = u.col.UpdateOne(c, bson.M{"_id": ID}, bson.M{"$push": bson.M{"keys": bson.M{"_id": key.ID, "ssh_hash": key.SSHHash, "name": key.Name, "created_at": key.CreatedAt}}})

	return err
}

func (u *userRepo) DeleteKey(c context.Context, id string) error {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{"keys._id": ID}
	update := bson.M{"$pull": bson.M{"keys": bson.M{"_id": ID}}}

	_, err = u.col.UpdateOne(c, filter, update)

	return err
}

func (u *userRepo) HasTheSameKey(c context.Context, id, str string) bool {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return true
	}

	mp := make(map[string]interface{}, 0)
	err = u.col.FindOne(c, bson.M{"_id": ID, "keys": bson.M{"$elemMatch": bson.M{"ssh_hash": str}}}).Decode(&mp)
	log.Println(mp)
	if err == nil {
		return true
	}
	if !errors.Is(err, mongo.ErrNoDocuments) && err != nil {
		return true
	}

	return false
}

func (u *userRepo) GetUserInfoByHashSSH(c context.Context, str string) (*User, error) {
	var user User
	err := u.col.FindOne(c, bson.M{"keys": bson.M{"$elemMatch": bson.M{"ssh_hash": str}}}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &user, nil
}

func (u *userRepo) GetUserInfoBySSH(c context.Context, str string) (*User, error) {
	var user User
	err := u.col.FindOne(c, bson.M{"keys": bson.M{"$elemMatch": bson.M{"ssh_key": str}}}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &user, nil
}
