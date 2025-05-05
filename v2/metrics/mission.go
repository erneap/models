package metrics

import (
	"strings"
	"time"

	"github.com/erneap/models/v2/systemdata"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MissionSensorOutage struct {
	TotalOutageMinutes     uint `json:"totalOutageMinutes" bson:"totalOutageMinutes"`
	PartialLBOutageMinutes uint `json:"partialLBOutageMinutes" bson:"partialLBOutageMinutes"`
	PartialHBOutageMinutes uint `json:"partialHBOutageMinutes" bson:"partialHBOutageMinutes"`
}

type MissionSensor struct {
	SensorID          string                  `json:"sensorID" bson:"sensorID"`
	SensorType        systemdata.GeneralTypes `json:"sensorType" bson:"sensorType"`
	PreflightMinutes  uint                    `json:"preflightMinutes" bson:"preflightMinutes"`
	ScheduledMinutes  uint                    `json:"scheduledMinutes" bson:"scheduledMinutes"`
	ExecutedMinutes   uint                    `json:"executedMinutes" bson:"executedMinutes"`
	PostflightMinutes uint                    `json:"postflightMinutes" bson:"postflightMinutes"`
	AdditionalMinutes uint                    `json:"additionalMinutes" bson:"additionalMinutes"`
	FinalCode         uint                    `json:"finalCode" bson:"finalCode"`
	KitNumber         string                  `json:"kitNumber" bson:"kitNumber"`
	SensorOutage      MissionSensorOutage     `json:"sensorOutage" bson:"sensorOutage"`
	GroundOutage      uint                    `json:"groundOutage" bson:"groundOutage"`
	HasHap            bool                    `json:"hasHap" bson:"hasHap"`
	TowerID           uint                    `json:"towerID,omitempty" bson:"towerID,omitempty"`
	SortID            uint                    `json:"sortID" bson:"sortID"`
	Comments          string                  `json:"comments" bson:"comments"`
	CheckedEquipment  []string                `json:"equipment,omitempty" bson:"equipment,omitempty"`
	Images            []systemdata.ImageType  `json:"images" bson:"images,omitempty"`
}

type ByMissionSensor []MissionSensor

func (c ByMissionSensor) Len() int { return len(c) }
func (c ByMissionSensor) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByMissionSensor) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (s *MissionSensor) EquipmentInUse(sid string) bool {
	answer := false
	if len(s.CheckedEquipment) > 0 {
		for _, s := range s.CheckedEquipment {
			if strings.EqualFold(s, sid) {
				answer = true
			}
		}
	} else {
		answer = true
	}
	return answer
}

type Mission struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	MissionDate    time.Time          `json:"missionDate" bson:"missionDate"`
	PlatformID     string             `json:"platformID" bson:"platformID"`
	SortieID       uint               `json:"sortieID" bson:"sortieID"`
	Exploitation   string             `json:"exploitation" bson:"exploitation"`
	TailNumber     string             `json:"tailNumber" bson:"tailNumber"`
	Communications string             `json:"communications" bson:"communications"`
	PrimaryDCGS    string             `json:"primaryDCGS" bson:"primaryDCGS"`
	Cancelled      bool               `json:"cancelled" bson:"cancelled"`
	Executed       bool               `json:"executed,omitempty" bson:"executed,omitempty"`
	Aborted        bool               `json:"aborted" bson:"aborted"`
	IndefDelay     bool               `json:"indefDelay" bson:"indefDelay"`
	MissionOverlap uint               `json:"missionOverlap" bson:"missionOverlap"`
	Comments       string             `json:"comments" bson:"comments"`
	Sensors        []MissionSensor    `json:"sensors,omitempty" bson:"sensors,omitempty"`
}

type ByMission []Mission

func (c ByMission) Len() int { return len(c) }
func (c ByMission) Less(i, j int) bool {
	if c[i].MissionDate.Equal(c[j].MissionDate) {
		if strings.EqualFold(c[i].PlatformID, c[j].PlatformID) {
			return c[i].SortieID < c[j].SortieID
		}
		return c[i].PlatformID < c[j].PlatformID
	}
	return c[i].MissionDate.Before(c[j].MissionDate)
}
func (c ByMission) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (m *Mission) EquipmentInUse(sid string) bool {
	answer := false
	if len(m.Sensors) > 0 {
		for _, s := range m.Sensors {
			if s.EquipmentInUse(sid) {
				answer = true
			}
		}
	} else {
		answer = true
	}
	return answer
}
