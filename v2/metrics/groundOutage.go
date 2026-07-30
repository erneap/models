package metrics

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type GroundOutage struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	OutageDate     time.Time          `json:"outageDate" bson:"outageDate"`
	GroundSystem   string             `json:"groundSystem" bson:"groundSystem"`
	Classification string             `json:"classification" bson:"classification"`
	OutageNumber   uint               `json:"outageNumber" bson:"outageNumber"`
	OutageMinutes  uint               `json:"outageMinutes" bson:"outageMinutes"`
	Subsystem      string             `json:"subSystem" bson:"subSystem"`
	ReferenceID    string             `json:"referenceId" bson:"referenceId"`
	MajorSystem    string             `json:"majorSystem" bson:"majorSystem"`
	Problem        string             `json:"problem" bson:"problem"`
	FixAction      string             `json:"fixAction" bson:"fixaction"`
	MissionOutage  bool               `json:"missionOutage" bson:"missionOutage"`
	Capability     string             `json:"capability,omitempty" bson:"capability,omitempty"`
}

type ByOutage []GroundOutage

func (c ByOutage) Len() int { return len(c) }
func (c ByOutage) Less(i, j int) bool {
	if c[i].OutageDate.Equal(c[j].OutageDate) {
		return c[i].OutageNumber < c[j].OutageNumber
	}
	return c[i].OutageDate.Before(c[j].OutageDate)
}
func (c ByOutage) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
