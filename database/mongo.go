package database

import (
	"chicko_chat/models"
	"context"
	"errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ChatDatabase struct {
	Users *mongo.Collection

	Messages *mongo.Collection

	Rooms *mongo.Collection
}


func (c *ChatDatabase) AddUser(user *data.UserData) (primitive.ObjectID, error) {

	res, err := c.Users.InsertOne(context.TODO(), user)

	if err != nil {
		return primitive.NilObjectID, err
	}

	return convertId(res)

}

// FindByEmail will be used to find a new user registry by email
func (c *ChatDatabase) FindByEmail(email string) (*data.UserData, error) {

	// Find user by email
	var user = data.UserData{}
	err := c.Users.FindOne(context.TODO(), bson.M{"email": email}).Decode(&user)

	if err != nil {
		// Checks if the user was not found
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil

}

func (c *ChatDatabase) CreateRoom(room *data.ChatRoom) (*data.ChatRoom, error) {

	res, err := c.Rooms.InsertOne(context.TODO(), room)

	if err != nil {
		return room, err
	}

	room.ID, err = convertId(res)

	if err != nil {
		return room, err
	}
	return room, nil

}


func (c *ChatDatabase) AddClientToRoom(room *data.ChatRoom, user *data.UserData) (*data.ChatRoom, error) {
	change := bson.M{
		"$push": bson.M{
			"users": user.ID,
		},
	}
	filter := bson.M{
		"_id": room.ID,
	}

	_, err := c.Rooms.UpdateOne(context.Background(), filter, change)
	if err != nil {
		return room, err
	}

	rm := &data.ChatRoom{}
	err = c.Rooms.FindOne(context.TODO(),filter).Decode(&rm)

	if err != nil {
		return room, err
	}
	return rm, nil

}

func (c *ChatDatabase) SaveMessage(message *data.ChatEvent) (primitive.ObjectID, error) {

	res, err := c.Messages.InsertOne(context.TODO(), message)

	if err != nil {
		return primitive.NilObjectID, err
	}

	return convertId(res)
}


// Get history of chat from the databse
func (c *ChatDatabase) GetHistoryOfRoom(room *data.ChatRoom) ([]data.ChatEvent, error) {

	findOptions := options.Find()
    cur, err := c.Messages.Find(context.TODO(), bson.D{{"roomid", room.ID}}, findOptions)
    if err != nil {
        return nil, err
    }
    defer cur.Close(context.TODO())

    var messages []data.ChatEvent
    err = cur.All(context.TODO(), &messages)

    return messages, nil
}