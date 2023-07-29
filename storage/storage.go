package storage

import (
	"github.com/SaidovZohid/swiftsend.it/storage/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
)

type StorageI interface {
	User() mongodb.UserI
	Session() mongodb.SessionI
	Usage() mongodb.UsageStorageI
}

type StoragePg struct {
	userRepo    mongodb.UserI
	sessionRepo mongodb.SessionI
	usageRepo   mongodb.UsageStorageI
}

func NewStorage(db *mongo.Database) StorageI {
	return &StoragePg{
		userRepo:    mongodb.NewUser(db),
		sessionRepo: mongodb.NewSession(db),
		usageRepo:   mongodb.NewUsage(db),
	}
}

func (s *StoragePg) User() mongodb.UserI {
	return s.userRepo
}

func (s *StoragePg) Session() mongodb.SessionI {
	return s.sessionRepo
}

func (s *StoragePg) Usage() mongodb.UsageStorageI {
	return s.usageRepo
}
