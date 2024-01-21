package hub

import (
	"context"
	"fmt"
	"github.com/dexguitar/wskfkchat/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type MongoRepo struct {
	client *mongo.Client
}

func NewMongoRepo(client *mongo.Client) *MongoRepo {
	return &MongoRepo{client: client}
}

func (m *MongoRepo) CreateRoom(ctx context.Context, id, name string) error {
	coll := m.client.Database(mongoDBName).Collection(mongoRoomColl)

	var room = &db.Room{}

	cur := coll.FindOne(ctx, bson.D{{"name", name}})
	err := cur.Decode(&room)
	if err != nil && err != mongo.ErrNoDocuments {
		return err
	}

	if room.Name != "" {
		return fmt.Errorf("room with name `%s` already exists", room.Name)
	}

	_, err = coll.InsertOne(ctx, bson.D{{"id", id}, {"name", name}})
	if err != nil {
		return fmt.Errorf("failed to insert document: %w", err)
	}

	return nil
}

func (m *MongoRepo) GetRooms(ctx context.Context) []db.Room {
	coll := m.client.Database(mongoDBName).Collection(mongoRoomColl)

	var rooms = make([]db.Room, 0)

	cur, err := coll.Find(ctx, bson.D{})
	if err != nil {
		return rooms
	}

	for cur.Next(ctx) {
		var res db.Room

		err = cur.Decode(&res)
		if err != nil {
			return rooms
		}

		rooms = append(rooms, res)
	}

	return rooms
}

func (m *MongoRepo) GetRoomByName(ctx context.Context, name string) (*db.Room, error) {
	coll := m.client.Database(mongoDBName).Collection(mongoRoomColl)

	var room = &db.Room{}

	res := coll.FindOne(ctx, bson.D{{"name", name}})
	err := res.Decode(room)
	if err == mongo.ErrNoDocuments {
		return room, fmt.Errorf("no room with name `%s`", name)
	}
	if err != nil {
		return room, err
	}

	return room, nil
}

func (m *MongoRepo) SaveMessage(ctx context.Context, message *Message) error {
	coll := m.client.Database("wskfkchat").Collection("messages")

	_, err := coll.InsertOne(ctx, message)
	if err != nil {
		return err
	}

	return nil
}
