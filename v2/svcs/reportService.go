package svcs

import (
	"context"
	"sort"
	"time"

	"github.com/erneap/models/config"
	"github.com/erneap/models/general"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func AddReport(typeid, subtype, mimetype string, body []byte) (*general.DBReport, error) {
	now := time.Now().UTC()
	oTypeID, err := primitive.ObjectIDFromHex(typeid)
	if err != nil {
		return nil, err
	}
	rpt := &general.DBReport{
		ID:            primitive.NewObjectID(),
		ReportDate:    now,
		ReportTypeID:  oTypeID,
		ReportSubType: subtype,
		MimeType:      mimetype,
	}
	rpt.SetDocument(body)

	rptCol := config.GetCollection(config.DB, "general", "reports")

	rptCol.InsertOne(context.TODO(), rpt)

	return rpt, nil
}

func AddReportWithDate(dt time.Time, typeid, subtype,
	mimetype string, body []byte) (*general.DBReport, error) {
	oTypeID, err := primitive.ObjectIDFromHex(typeid)
	if err != nil {
		return nil, err
	}
	rpt := &general.DBReport{
		ID:            primitive.NewObjectID(),
		ReportDate:    dt,
		ReportTypeID:  oTypeID,
		ReportSubType: subtype,
		MimeType:      mimetype,
	}
	rpt.SetDocument(body)

	rptCol := config.GetCollection(config.DB, "general", "reports")

	rptCol.InsertOne(context.TODO(), rpt)

	return rpt, nil
}

func UpdateReport(id, mimetype string, body []byte) (*general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var rpt *general.DBReport
	err = rptCol.FindOne(context.TODO(), filter).Decode(&rpt)
	if err != nil {
		return nil, err
	}

	rpt.ReportDate = time.Now().UTC()
	rpt.MimeType = mimetype
	rpt.SetDocument(body)
	_, err = rptCol.ReplaceOne(context.TODO(), filter, rpt)
	if err != nil {
		return nil, err
	}
	return rpt, nil
}

func DeleteReport(id string) error {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": oId,
	}

	_, err = rptCol.DeleteOne(context.TODO(), filter)
	return err
}

func PurgeReports(dt time.Time) error {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	filter := bson.M{
		"reportdate": bson.M{"$lt": dt},
	}

	_, err := rptCol.DeleteMany(context.TODO(), filter)
	return err
}

func GetReport(id string) (*general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}

	var rpt *general.DBReport
	err = rptCol.FindOne(context.TODO(), filter).Decode(&rpt)
	if err != nil {
		return nil, err
	}
	return rpt, nil
}

func GetReportsByType(id string) ([]general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")
	oTypeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"reporttypeid": oTypeID,
	}

	var rpts []general.DBReport

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByDBReports(rpts))
	return rpts, nil
}

func GetReportsBetweenDates(date1, date2 time.Time) ([]general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	filter := bson.M{
		"reportdate": bson.M{"$gte": primitive.NewDateTimeFromTime(date1),
			"$lte": primitive.NewDateTimeFromTime(date2.AddDate(0, 0, 1))},
	}

	var rpts []general.DBReport

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByDBReports(rpts))
	return rpts, nil
}

func GetReportsByTypeAndDates(id string, date1, date2 time.Time) ([]general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")
	oTypeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	filter := bson.M{
		"reporttypeid": oTypeID,
		"reportdate": bson.M{"$gte": primitive.NewDateTimeFromTime(date1),
			"$lt": primitive.NewDateTimeFromTime(date2.AddDate(0, 0, 1))},
	}

	var rpts []general.DBReport

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByDBReports(rpts))
	return rpts, nil
}

func GetReportsAll() ([]general.DBReport, error) {
	rptCol := config.GetCollection(config.DB, "general", "reports")

	filter := bson.M{}

	var rpts []general.DBReport

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByDBReports(rpts))
	return rpts, nil
}

// CRUD methods for report types
func CreateReportType(app, name, rpttype string, subtypes []string) (*general.ReportType, error) {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")

	rpt := &general.ReportType{
		ID:             primitive.NewObjectID(),
		Application:    app,
		ReportTypeName: name,
		ReportType:     rpttype,
		SubTypes:       subtypes,
	}

	_, err := rptCol.InsertOne(context.TODO(), rpt)
	if err != nil {
		return nil, err
	}

	return rpt, nil
}

func UpdateReportType(id, app, name, rpttype string,
	subtypes []string) (*general.ReportType, error) {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")

	oID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oID,
	}

	var rpt general.ReportType
	err = rptCol.FindOne(context.TODO(), filter).Decode(&rpt)
	if err != nil {
		return nil, err
	}

	rpt.Application = app
	rpt.ReportTypeName = name
	rpt.ReportType = rpttype
	rpt.SubTypes = subtypes

	_, err = rptCol.ReplaceOne(context.TODO(), filter, rpt)
	if err != nil {
		return nil, err
	}

	return &rpt, nil
}

func DeleteReportType(id string) error {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")

	oID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}

	filter := bson.M{
		"_id": oID,
	}

	_, err = rptCol.DeleteOne(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}

func GetReportTypes() ([]general.ReportType, error) {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")

	filter := bson.M{}

	var rpts []general.ReportType

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByReportTypes(rpts))
	return rpts, nil

}

func GetReportTypesByApplication(app string) ([]general.ReportType, error) {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")

	filter := bson.M{
		"application": app,
	}

	var rpts []general.ReportType

	cursor, err := rptCol.Find(context.TODO(), filter)
	if err != nil {
		return rpts, err
	}

	if err = cursor.All(context.TODO(), &rpts); err != nil {
		return rpts, err
	}
	sort.Sort(general.ByReportTypes(rpts))
	return rpts, nil

}

func GetReportType(id string) (*general.ReportType, error) {
	rptCol := config.GetCollection(config.DB, "general", "reporttypes")
	oId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}

	filter := bson.M{
		"_id": oId,
	}
	var rpt general.ReportType
	err = rptCol.FindOne(context.TODO(), filter).Decode(&rpt)
	if err != nil {
		return nil, err
	}
	return &rpt, nil
}
