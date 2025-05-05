package svcs

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/erneap/go-models/config"
	"github.com/erneap/go-models/notifications"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// notification retrieve functions (All, by Employee, and single)

func GetAllMessages() ([]notifications.Notification, error) {
	noteCol := config.GetCollection(config.DB, "scheduler", "notifications")

	var list []notifications.Notification

	cursor, err := noteCol.Find(context.TODO(), bson.M{})
	if err != nil {
		return list, err
	}

	if err = cursor.All(context.TODO(), &list); err != nil {
		return list, err
	}

	sort.Sort(notifications.ByNofication(list))

	return list, nil
}

func GetMessagesByEmployee(id string) ([]notifications.Notification, error) {
	noteCol := config.GetCollection(config.DB, "scheduler", "notifications")

	var list []notifications.Notification

	filter := bson.M{
		"to": id,
	}

	cursor, err := noteCol.Find(context.TODO(), filter)
	if err != nil {
		return list, err
	}

	if err = cursor.All(context.TODO(), &list); err != nil {
		return list, err
	}

	sort.Sort(notifications.ByNofication(list))

	return list, nil
}

func GetMessage(id string) (notifications.Notification, error) {
	noteCol := config.GetCollection(config.DB, "scheduler", "notifications")

	var answer notifications.Notification
	mid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return answer, err
	}

	filter := bson.M{
		"_id": mid,
	}

	err = noteCol.FindOne(context.TODO(), filter).Decode(&answer)
	if err != nil {
		return answer, err
	}

	return answer, nil
}

// Create function which include receipent, sender and message.
// the identifier and date are automatic.
func CreateMessage(to, from, message string) error {
	noteCol := config.GetCollection(config.DB, "scheduler", "notifications")

	msg := &notifications.Notification{
		ID:      primitive.NewObjectID(),
		Date:    time.Now().UTC(),
		To:      to,
		From:    from,
		Message: message,
	}

	result, err := noteCol.InsertOne(context.TODO(), msg)
	if err != nil {
		return err
	}
	if result.InsertedID == primitive.NilObjectID {
		return errors.New("not created")
	}
	return nil
}

// There is no update routine because messages can't be updated manually.

// After the message is viewed, it will be acknowledged and removed
// from the database.  This is the only delete routine.
func DeleteMessage(id string) error {
	noteCol := config.GetCollection(config.DB, "scheduler", "notifications")

	mid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": mid,
	}

	result, err := noteCol.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	if result.DeletedCount <= 0 {
		return errors.New("no message deleted")
	}

	return nil
}
