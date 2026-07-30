package systemdata

type GroundSystemExploitation struct {
	PlatformID      string `json:"platformID" bson:"platformID"`
	SensorType      string `json:"sensorType" bson:"sensorType"`
	Exploitation    string `json:"exploitation" bson:"exploitation"`
	CommunicationID string `json:"communicationID" bson:"communicationID"`
}

type ByGSExploitation []GroundSystemExploitation

func (c ByGSExploitation) Len() int { return len(c) }
func (c ByGSExploitation) Less(i, j int) bool {
	if c[i].PlatformID == c[j].PlatformID {
		if c[i].SensorType == c[j].SensorType {
			if c[i].Exploitation == c[j].Exploitation {
				return c[i].CommunicationID < c[j].CommunicationID
			}
			return c[i].Exploitation < c[j].Exploitation
		}
		return c[i].SensorType < c[j].SensorType
	}
	return c[i].PlatformID < c[j].PlatformID
}
func (c ByGSExploitation) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type GroundSystem struct {
	ID            string                     `json:"id" bson:"id"`
	Enclaves      []string                   `json:"enclaves" bson:"enclaves"`
	ShowOnGEOINT  bool                       `json:"showOnGEOINT" bson:"showOnGEOINT"`
	ShowOnGSEG    bool                       `json:"showOnGSEG" bson:"showOnGSEG"`
	ShowOnMIST    bool                       `json:"showOnMIST" bson:"showOnMIST"`
	ShowOnXINT    bool                       `json:"showOnXINT" bson:"showOnXINT"`
	CheckForUse   bool                       `json:"checkForUse,omitempty" bson:"checkForUse,omitempty"`
	Exploitations []GroundSystemExploitation `json:"exploitations" bson:"exploitations"`
}

type ByGroundSystem []GroundSystem

func (c ByGroundSystem) Len() int { return len(c) }
func (c ByGroundSystem) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByGroundSystem) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (gs *GroundSystem) UseMissionSensor(platform, sensor, exploit, comm string) bool {
	if len(gs.Exploitations) == 0 {
		return true
	}
	answer := false
	for _, exp := range gs.Exploitations {
		if exp.PlatformID == platform && exp.SensorType == sensor &&
			exp.Exploitation == exploit && exp.CommunicationID == comm {
			answer = true
		}
	}
	return answer
}
