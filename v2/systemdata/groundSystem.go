package systemdata

import "strings"

type GroundSystemExploitation struct {
	PlatformID      string   `json:"platformID" bson:"platformID"`
	SensorType      string   `json:"sensorType" bson:"sensorType"`
	Exploitation    string   `json:"exploitation" bson:"exploitation"`
	CommunicationID string   `json:"communicationID" bson:"communicationID"`
	Enclaves        []string `json:"enclaves,omitempty" bson:"enclaves,omitempty"`
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

func (gs *GroundSystem) UseMissionSensor(platform, sensor, exploit, comm, enclave string) bool {
	if len(gs.Exploitations) == 0 {
		return true
	}
	sensor = strings.ToLower(sensor)
	exploit = strings.ToLower(exploit)
	comm = strings.ToLower(comm)
	answer := false
	for _, exp := range gs.Exploitations {
		if strings.EqualFold(exp.PlatformID, platform) &&
			strings.Contains(strings.ToLower(exp.SensorType), sensor) &&
			strings.Contains(strings.ToLower(exp.Exploitation), exploit) &&
			strings.Contains(strings.ToLower(exp.CommunicationID), comm) {
			if enclave != "" {
				if len(exp.Enclaves) > 0 {
					for _, enc := range exp.Enclaves {
						if strings.EqualFold(enclave, enc) {
							answer = true
						}
					}
				} else {
					answer = true
				}
			} else {
				answer = true
			}
		}
	}
	return answer
}
