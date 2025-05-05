package systemdata

import (
	"strings"
)

type SensorStandardTimes struct {
	PreflightMinutes  uint `json:"preflightMinutes" bson:"preflightMinutes"`
	ScheduledMinutes  uint `json:"scheduledMinutes" bson:"scheduledMinutes"`
	PostflightMinutes uint `json:"postflightMinutes" bson:"postflightMinutes"`
}

type SensorExploitation struct {
	Exploitation  string              `json:"exploitation" bson:"exploitation"`
	ShowOnGEOINT  bool                `json:"showOnGEOINT" bson:"showOnGEOINT"`
	ShowOnGSEG    bool                `json:"showOnGSEG" bson:"showOnGSEG"`
	ShowOnMIST    bool                `json:"showOnMIST" bson:"showOnMIST"`
	ShowOnXINT    bool                `json:"showOnXINT" bson:"showOnXINT"`
	StandardTimes SensorStandardTimes `json:"standardTimes" bson:"standardTimes"`
}

type BySensorExploitations []SensorExploitation

func (c BySensorExploitations) Len() int { return len(c) }
func (c BySensorExploitations) Less(i, j int) bool {
	return c[i].Exploitation < c[j].Exploitation
}
func (c BySensorExploitations) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type GeneralTypes int

const (
	GEOINT GeneralTypes = 1
	XINT   GeneralTypes = 2
	MIST   GeneralTypes = 3
	SYERS  GeneralTypes = 4
	ADMIN  GeneralTypes = 9
	OTHER  GeneralTypes = 99
	ALL    GeneralTypes = 9999
)

type PlatformSensor struct {
	ID             string               `json:"id"`
	Association    string               `json:"association"`
	GeneralType    GeneralTypes         `json:"generalType"`
	ShowTailNumber bool                 `json:"showTailNumber"`
	SortID         uint                 `json:"sortID"`
	Exploitations  []SensorExploitation `json:"exploitations,omitempty"`
	ImageTypes     []ImageType          `json:"imageTypes,omitempty"`
}

type ByPlatformSensor []PlatformSensor

func (c ByPlatformSensor) Len() int { return len(c) }
func (c ByPlatformSensor) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByPlatformSensor) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (ps *PlatformSensor) UseForExploitation(exploit string, rpt GeneralTypes) bool {
	if rpt <= 0 || rpt > SYERS {
		rpt = ALL
	}
	answer := false
	for _, exp := range ps.Exploitations {
		use := strings.Contains(strings.ToLower(exp.Exploitation),
			strings.ToLower(exploit))
		if use &&
			(rpt == ALL || (rpt == XINT && exp.ShowOnXINT) ||
				(rpt == GEOINT && exp.ShowOnGEOINT) || (rpt == SYERS && exp.ShowOnGSEG) ||
				(rpt == MIST && exp.ShowOnMIST)) {
			answer = true
		}
	}
	return answer
}

type Platform struct {
	ID      string           `json:"id"`
	Sensors []PlatformSensor `json:"sensors"`
	SortID  uint             `json:"sortID"`
}

type ByPlatform []Platform

func (c ByPlatform) Len() int { return len(c) }
func (c ByPlatform) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByPlatform) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p *Platform) ShowOnSummary(exploit string, rpt GeneralTypes) bool {
	if rpt < GEOINT || rpt > SYERS {
		rpt = ALL
	}
	answer := false
	for _, sen := range p.Sensors {
		if sen.UseForExploitation(exploit, rpt) {
			answer = true
		}
	}
	return answer
}
