package mongodb

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type User struct {
	Id                     primitive.ObjectID `bson:"_id"`
	Username               string             `json:"username"`
	Fullname               string             `bson:"fullname"`
	Email                  *string            `bson:"email"`
	Subdomain              *string            `bson:"subdomain"`
	CreatedAt              string             `bson:"created_at"`
	LogInAndSignUpProvider string             `bson:"provider"` // github or google
	Keys                   []Keys             `bson:"keys"`
}

type userRepo struct {
	col *mongo.Collection
}

type UserI interface {
	RegisterUserFirst(c context.Context, user *User) (interface{}, error)
	FindUserByUsername(c context.Context, username string) (*User, error)
	FindUserByUsernameForLS(c context.Context, username, provider string) (*User, error)
	FindUserByEmail(c context.Context, email string) (*User, error)
	FindUserBySubdomain(c context.Context, subdomain string) (*User, error)
	SetSubdomainAndSShKey(c context.Context, id, subdomain string, key *Keys) error
	PushNewKey(c context.Context, id string, key *Keys) error
	DeleteKey(c context.Context, id string) error
	HasTheSameKey(c context.Context, id, str string) bool
	FindUserByID(c context.Context, id string) (*User, error)
	GetUserInfoBySSH(c context.Context, str string) (*User, error)
	GetUserInfoByHashSSH(c context.Context, str string) (*User, error)
	GetAllUsers(c context.Context) ([]User, error)
}

func NewUser(db *mongo.Database) UserI {
	return &userRepo{
		col: db.Collection("users"),
	}
}

func (u *userRepo) RegisterUserFirst(c context.Context, user *User) (interface{}, error) {
	user.Id = primitive.NewObjectID()
	user.Keys = []Keys{}

	res, err := u.col.InsertOne(c, user)
	if err != nil {
		return nil, err
	}

	return res.InsertedID, nil
}

func (u *userRepo) FindUserByUsername(c context.Context, username string) (*User, error) {
	var res User

	err := u.col.FindOne(c, bson.M{"username": username}).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &res, nil
}

func (u *userRepo) FindUserByID(c context.Context, id string) (*User, error) {
	ID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	var res User

	err = u.col.FindOne(c, bson.M{"_id": ID}).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &res, nil
}

func (u *userRepo) FindUserByUsernameForLS(c context.Context, username, provider string) (*User, error) {
	var res User

	err := u.col.FindOne(c, bson.M{"username": username, "provider": provider}).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {

			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &res, nil
}

func (u *userRepo) FindUserBySubdomain(c context.Context, subdomain string) (*User, error) {
	var user User
	err := u.col.FindOne(c, bson.M{"subdomain": subdomain}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &user, nil
}

func (u *userRepo) FindUserByEmail(c context.Context, email string) (*User, error) {
	var user User
	err := u.col.FindOne(c, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, mongo.ErrNoDocuments
		}
		return nil, err
	}

	return &user, nil
}

func (u *userRepo) GetAllUsers(c context.Context) ([]User, error) {
	filter := bson.D{}

	// Find all users from the collection
	cur, err := u.col.Find(c, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	// Create a slice to hold the users
	users := make([]User, 0)

	// Iterate through the cursor and decode each user into the users slice
	for cur.Next(c) {
		var user User
		if err := cur.Decode(&user); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	// Check for any errors during the iteration
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return users, nil
}
