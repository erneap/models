package entries

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserPreferences struct {
	ReadingPlanID primitive.ObjectID `json:"readingplan" bson:"readingplan"`
	BibleVersion  string             `json:"bibleversion" bson:"bibleversion"`
	FontSize      string             `json:"fontsize" bson:"fontsize"`
	Language      string             `json:"language" bson:"language"`
}
