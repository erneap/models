package svcs

import (
	"context"
	"sort"
	"time"

	"github.com/erneap/go-models/config"
	"github.com/erneap/go-models/general"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateDBLogEntryWithDate(dt time.Time, app, cat, title, name, msg string) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	// new log entry
	entry := &general.LogEntry{
		ID:          primitive.NewObjectID(),
		EntryDate:   dt,
		Application: app,
		Category:    cat,
		Title:       title,
		Name:        name,
		Message:     msg,
	}

	_, err := logCol.InsertOne(context.TODO(), entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

// CRUD Methods for this data collection
func CreateDBLogEntry(app, cat, title, name, msg string, c *gin.Context) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	if name == "" && c != nil {
		userid := GetRequestor(c)
		if userid != "" {
			user, _ := GetUserByID(userid)
			if user != nil {
				name = user.LastName
			}
		}
	}

	// new log entry
	entry := &general.LogEntry{
		ID:          primitive.NewObjectID(),
		EntryDate:   time.Now().UTC(),
		Application: app,
		Category:    cat,
		Title:       title,
		Name:        name,
		Message:     msg,
	}

	_, err := logCol.InsertOne(context.TODO(), entry)
	if err != nil {
		return nil, err
	}
	return entry, nil
}

func UpdateDBLogEntry(id, cat, title, name, msg string) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var entry general.LogEntry
	err = logCol.FindOne(context.TODO(), filter).Decode(&entry)
	if err != nil {
		return nil, err
	}

	if cat != "" {
		entry.Category = cat
	}
	if title != "" {
		entry.Title = title
	}
	if name != "" {
		entry.Name = name
	}
	entry.Message = msg

	_, err = logCol.ReplaceOne(context.TODO(), filter, &entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func DeleteDBLogEntry(id string) error {
	logCol := config.GetCollection(config.DB, "general", "logs")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": oId,
	}

	_, err = logCol.DeleteOne(context.TODO(), filter)
	return err
}

func PurgeLogs(dt time.Time) error {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{
		"entrydate": bson.M{"$lt": dt},
	}

	_, err := logCol.DeleteMany(context.TODO(), filter)
	return err
}

func GetDBLogEntry(id string) (*general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var entry general.LogEntry
	err = logCol.FindOne(context.TODO(), filter).Decode(&entry)
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

func GetDBLogEntriesAll() ([]general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{}

	var logs []general.LogEntry

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}
	sort.Sort(general.ByLogEntries(logs))
	return logs, nil
}

func GetDBLogEntriesByApplication(app string) ([]general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{
		"application": app,
	}

	var logs []general.LogEntry

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}
	sort.Sort(general.ByLogEntries(logs))
	return logs, nil
}

func GetDBLogEntriesByApplicationBetweenDates(app string, dt1,
	dt2 time.Time) ([]general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{
		"application": app,
		"entrydate": bson.M{
			"$gte": dt1,
			"$lte": dt2,
		},
	}

	var logs []general.LogEntry

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}
	sort.Sort(general.ByLogEntries(logs))
	return logs, nil
}

func GetDBLogEntriesByApplicationCategory(app, cat string) ([]general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{
		"application": app,
		"category":    cat,
	}

	var logs []general.LogEntry

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}
	sort.Sort(general.ByLogEntries(logs))
	return logs, nil
}

func GetDBLogEntriesByApplicationCategoryBetweenDates(app, cat string, dt1,
	dt2 time.Time) ([]general.LogEntry, error) {
	logCol := config.GetCollection(config.DB, "general", "logs")

	filter := bson.M{
		"application": app,
		"category":    cat,
		"entrydate": bson.M{
			"$gte": dt1,
			"$lte": dt2,
		},
	}

	var logs []general.LogEntry

	cursor, err := logCol.Find(context.TODO(), filter)
	if err != nil {
		return logs, err
	}

	if err = cursor.All(context.TODO(), &logs); err != nil {
		return logs, err
	}
	sort.Sort(general.ByLogEntries(logs))
	return logs, nil
}
