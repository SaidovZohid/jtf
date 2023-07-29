package mongodb

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type Session struct {
	SessionID   primitive.ObjectID `bson:"_id"`
	UserID      string             `bson:"user_id"`
	AccessToken string             `bson:"access_token"`
	IpAddress   string             `bson:"ip_address"`
	Device      string             `bson:"device"`
	Timezone    string             `bson:"timezone"`
	LastLogin   string             `bson:"last_login"`
	CreatedAt   string             `bson:"created_at"`
}

type session struct {
	col *mongo.Collection
}

type SessionI interface {
	CreateSession(c context.Context, s *Session) (string, error)
	GetSessionByID(c context.Context, sessionID string) (*Session, error)
	DeleteSessionByID(c context.Context, sessionID string) error
	GetAllSessions(c context.Context) ([]Session, error)
}

func NewSession(db *mongo.Database) SessionI {
	return &session{
		col: db.Collection("sessions"),
	}
}

func (ses *session) CreateSession(c context.Context, s *Session) (string, error) {
	timezone := time.FixedZone("GMT+5", 5*60*60) // 5 hours ahead of UTC
	isExist, err := ses.GetSession(c, s.UserID, s.IpAddress, s.Device)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", err
	}
	if isExist == nil {
		s.SessionID = primitive.NewObjectID()
		s.CreatedAt = time.Now().In(timezone).Format(time.RFC1123)
		s.LastLogin = time.Now().In(timezone).Format(time.RFC1123)
		res, err := ses.col.InsertOne(c, s)
		if err != nil {
			return "", err
		}

		return res.InsertedID.(primitive.ObjectID).Hex(), nil
	}

	_, err = ses.col.UpdateOne(c, bson.M{"_id": isExist.SessionID}, bson.M{"$set": bson.M{"access_token": s.AccessToken}, "last_login": time.Now().In(timezone).Format(time.RFC1123)})

	return isExist.SessionID.Hex(), err
}

func (ses *session) GetSessionByID(c context.Context, sessionID string) (*Session, error) {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return nil, err
	}

	var res Session

	err = ses.col.FindOne(c, bson.M{"_id": id}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ses *session) GetSession(c context.Context, userId, ipAddress, device string) (*Session, error) {
	var res Session

	err := ses.col.FindOne(c, bson.M{"user_id": userId, "ip_address": ipAddress, "device": device}).Decode(&res)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (ses *session) DeleteSessionByID(c context.Context, sessionID string) error {
	id, err := primitive.ObjectIDFromHex(sessionID)
	if err != nil {
		return err
	}

	_, err = ses.col.DeleteOne(c, bson.M{"_id": id})
	return err
}

func (u *session) GetAllSessions(c context.Context) ([]Session, error) {
	filter := bson.D{}

	// Find all users from the collection
	cur, err := u.col.Find(c, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(c)

	// Create a slice to hold the users
	sessions := make([]Session, 0)

	// Iterate through the cursor and decode each user into the users slice
	for cur.Next(c) {
		var session Session
		if err := cur.Decode(&session); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	// Check for any errors during the iteration
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return sessions, nil
}
