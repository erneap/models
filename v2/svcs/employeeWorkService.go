package svcs

import (
	"context"
	"log"
	"time"

	"github.com/erneap/models/v2/config"
	"github.com/erneap/models/v2/employees"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Employee Work is storage records for work accomplished by an employee during
// a single year.
// Every service will have functions for completing the CRUD functions
// the retrieve function will only be for individual employee's year

func CreateEmployeeWork(work *employees.EmployeeWorkRecord) error {
	empWCol := config.GetCollection(config.DB, "scheduler", "employeework")

	filter := bson.M{
		"employeeID": work.EmployeeID,
		"year":       work.Year,
	}

	var tEmpWork employees.EmployeeWorkRecord

	err := empWCol.FindOne(context.TODO(), filter).Decode(&tEmpWork)
	if err == mongo.ErrNoDocuments {
		work.ID = primitive.NewObjectID()
		_, err = empWCol.InsertOne(context.TODO(), work)
		return err
	} else {
		filter = bson.M{
			"_id": tEmpWork.ID,
		}
		work.ID = tEmpWork.ID
		_, err = empWCol.ReplaceOne(context.TODO(), filter, work)
		return err
	}
}

func GetEmployeeWork(id string, year uint) (*employees.EmployeeWorkRecord, error) {
	empWCol := config.GetCollection(config.DB, "scheduler", "employeework")

	empID, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{
		"employeeID": empID,
		"year":       year,
	}

	var eWork employees.EmployeeWorkRecord

	err := empWCol.FindOne(context.TODO(), filter).Decode(&eWork)
	if err != nil {
		return nil, err
	}
	return &eWork, nil
}

func UpdateEmployeeWork(eWork *employees.EmployeeWorkRecord) error {
	empWCol := config.GetCollection(config.DB, "scheduler", "employeework")

	filter := bson.M{
		"_id": eWork.ID,
	}

	_, err := empWCol.ReplaceOne(context.TODO(), filter, eWork)
	return err
}

func DeleteEmployeeWork(id string, year uint) error {
	empWCol := config.GetCollection(config.DB, "scheduler", "employeework")

	empID, _ := primitive.ObjectIDFromHex(id)

	filter := bson.M{
		"employeeID": empID,
		"year":       year,
	}

	_, err := empWCol.DeleteOne(context.TODO(), filter)
	return err
}

func GetEmployeeWorkForPurge(purgeDate time.Time) ([]employees.EmployeeWorkRecord, error) {
	empWCol := config.GetCollection(config.DB, "scheduler", "employeework")

	filter := bson.M{"year": bson.M{"$lte": purgeDate.Year()}}

	var empWork []employees.EmployeeWorkRecord

	cursor, err := empWCol.Find(context.TODO(), filter)
	if err != nil {
		return empWork[:0], err
	}

	if err = cursor.All(context.TODO(), &empWork); err != nil {
		log.Println(err)
	}
	return empWork, err
}
