package notifications

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Notification struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	Date    time.Time          `json:"date" bson:"date"`
	To      string             `json:"to" bson:"to"`
	From    string             `json:"from" bson:"from"`
	Message string             `json:"message" bson:"bson"`
}

type ByNofication []Notification

func (c ByNofication) Len() int { return len(c) }
func (c ByNofication) Less(i, j int) bool {
	return c[i].Date.Before(c[j].Date)
}
func (c ByNofication) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Message struct {
	Message string `json:"message"`
}
