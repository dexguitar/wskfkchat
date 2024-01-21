package db

type Room struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`
}
