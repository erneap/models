package general

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type LogEntry struct {
	ID          primitive.ObjectID `json:"id" bson:"_id"`
	EntryDate   time.Time          `json:"entrydate" bson:"entrydate"`
	Application string             `json:"application" bson:"application"`
	Category    string             `json:"category" bson:"category"`
	Title       string             `json:"title" bson:"title"`
	Name        string             `json:"name" bson:"name"`
	Message     string             `json:"message" bson:"message"`
}

type ByLogEntries []LogEntry

func (c ByLogEntries) Len() int { return len(c) }
func (c ByLogEntries) Less(i, j int) bool {
	if c[i].Application == c[j].Application {
		return c[i].EntryDate.Before(c[j].EntryDate)
	}
	return c[i].Application < c[j].Application
}
func (c ByLogEntries) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
