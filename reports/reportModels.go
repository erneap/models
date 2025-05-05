package reports

import (
	"strings"
	"time"

	"github.com/erneap/go-models/metrics"
	"github.com/erneap/go-models/systemdata"
)

type MissionDay struct {
	MissionDate time.Time
	Missions    []metrics.Mission
}

type ByMissionDay []MissionDay

func (c ByMissionDay) Len() int { return len(c) }
func (c ByMissionDay) Less(i, j int) bool {
	return c[i].MissionDate.Before(c[j].MissionDate)
}
func (c ByMissionDay) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type MissionType struct {
	Exploitation string
	Platform     string
	Missions     []metrics.Mission
}

func (mt *MissionType) GetScheduled(exploit string, sens []string) int {
	answer := 0
	if (strings.ToLower(exploit) == "primary" &&
		strings.ToLower(mt.Exploitation) == "primary") ||
		(strings.ToLower(exploit) != "primary" &&
			strings.ToLower(mt.Exploitation) != "primary") {
		if len(sens) == 0 {
			answer = len(mt.Missions)
		} else {
			for _, msn := range mt.Missions {
				found := false
				msn.Decrypt()
				for _, sen := range msn.MissionData.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) {
							if !found {
								answer++
								found = true
							}
						}
					}
				}
			}
		}
	}
	return answer
}

func (mt *MissionType) GetExecuted(exploit string, sens []string) int {
	answer := 0
	if (strings.ToLower(exploit) == "primary" &&
		strings.ToLower(mt.Exploitation) == "primary") ||
		(strings.ToLower(exploit) != "primary" &&
			strings.ToLower(mt.Exploitation) != "primary") {
		if len(sens) == 0 {
			for _, msn := range mt.Missions {
				msn.Decrypt()
				if !msn.MissionData.Aborted && !msn.MissionData.Cancelled &&
					!msn.MissionData.IndefDelay {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				msn.Decrypt()
				if !msn.MissionData.Aborted && !msn.MissionData.Cancelled &&
					!msn.MissionData.IndefDelay {
					for _, sen := range msn.MissionData.Sensors {
						for _, lsen := range sens {
							if strings.EqualFold(sen.SensorID, lsen) {
								if !found {
									answer++
									found = true
								}
							}
						}
					}
				}
			}
		}
	}
	return answer
}

func (mt *MissionType) GetCancelled(exploit string, sens []string) int {
	answer := 0
	if (strings.ToLower(exploit) == "primary" &&
		strings.ToLower(mt.Exploitation) == "primary") ||
		(strings.ToLower(exploit) != "primary" &&
			strings.ToLower(mt.Exploitation) != "primary") {
		if len(sens) == 0 {
			for _, msn := range mt.Missions {
				msn.Decrypt()
				if msn.MissionData.Cancelled || msn.MissionData.IndefDelay {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				msn.Decrypt()
				if msn.MissionData.Cancelled || msn.MissionData.IndefDelay {
					for _, sen := range msn.MissionData.Sensors {
						for _, lsen := range sens {
							if strings.EqualFold(sen.SensorID, lsen) {
								if !found {
									answer++
									found = true
								}
							}
						}
					}
				}
			}
		}
	}
	return answer
}

func (mt *MissionType) GetAborted(exploit string, sens []string) int {
	answer := 0
	if (strings.ToLower(exploit) == "primary" &&
		strings.ToLower(mt.Exploitation) == "primary") ||
		(strings.ToLower(exploit) != "primary" &&
			strings.ToLower(mt.Exploitation) != "primary") {
		if len(sens) == 0 {
			for _, msn := range mt.Missions {
				msn.Decrypt()
				if msn.MissionData.Aborted {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				msn.Decrypt()
				if msn.MissionData.Aborted {
					for _, sen := range msn.MissionData.Sensors {
						for _, lsen := range sens {
							if strings.EqualFold(sen.SensorID, lsen) {
								if !found {
									answer++
									found = true
								}
							}
						}
					}
				}
			}
		}
	}
	return answer
}

func (mt *MissionType) GetPremissionTime(sens []string,
	gs *systemdata.GroundSystem) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		msn.Decrypt()
		msnExp := strings.ToLower(msn.MissionData.Exploitation)
		msnComm := strings.ToLower(msn.MissionData.Communications)
		if gs != nil {
			for _, exp := range gs.Exploitations {
				gsExp := strings.ToLower(exp.Exploitation)
				gsComm := strings.ToLower(exp.CommunicationID)
				if strings.EqualFold(msn.PlatformID, exp.PlatformID) &&
					strings.Contains(gsExp, msnExp) &&
					strings.Contains(gsComm, msnComm) {
					for _, mSen := range msn.MissionData.Sensors {
						if strings.EqualFold(exp.SensorType, mSen.SensorID) &&
							senMax < mSen.PreflightMinutes {
							senMax = mSen.PreflightMinutes
						}
					}
				}
			}
		} else {
			for _, sen := range msn.MissionData.Sensors {
				for _, lsen := range sens {
					if strings.EqualFold(sen.SensorID, lsen) &&
						senMax < sen.PreflightMinutes {
						senMax = sen.PreflightMinutes
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetPostmissionTime(sens []string,
	gs *systemdata.GroundSystem) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		msn.Decrypt()
		msnExp := strings.ToLower(msn.MissionData.Exploitation)
		msnComm := strings.ToLower(msn.MissionData.Communications)
		if gs != nil {
			for _, exp := range gs.Exploitations {
				gsExp := strings.ToLower(exp.Exploitation)
				gsComm := strings.ToLower(exp.CommunicationID)
				if strings.EqualFold(msn.PlatformID, exp.PlatformID) &&
					strings.Contains(gsExp, msnExp) &&
					strings.Contains(gsComm, msnComm) {
					for _, mSen := range msn.MissionData.Sensors {
						if strings.EqualFold(exp.SensorType, mSen.SensorID) &&
							senMax < mSen.PostflightMinutes {
							senMax = mSen.PostflightMinutes
						}
					}
				}
			}
		} else {
			for _, sen := range msn.MissionData.Sensors {
				for _, lsen := range sens {
					if strings.EqualFold(sen.SensorID, lsen) &&
						senMax < sen.PostflightMinutes {
						senMax = sen.PostflightMinutes
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetScheduledTime(sens []string,
	gs *systemdata.GroundSystem) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		msn.Decrypt()
		msnExp := strings.ToLower(msn.MissionData.Exploitation)
		msnComm := strings.ToLower(msn.MissionData.Communications)
		if gs != nil {
			for _, exp := range gs.Exploitations {
				gsExp := strings.ToLower(exp.Exploitation)
				gsComm := strings.ToLower(exp.CommunicationID)
				if strings.EqualFold(msn.PlatformID, exp.PlatformID) &&
					strings.Contains(gsExp, msnExp) &&
					strings.Contains(gsComm, msnComm) {
					for _, mSen := range msn.MissionData.Sensors {
						if strings.EqualFold(exp.SensorType, mSen.SensorID) &&
							senMax < mSen.ScheduledMinutes {
							senMax = mSen.ScheduledMinutes
						}
					}
				}
			}
		} else {
			for _, sen := range msn.MissionData.Sensors {
				for _, lsen := range sens {
					if strings.EqualFold(sen.SensorID, lsen) &&
						senMax < sen.ScheduledMinutes {
						senMax = sen.ScheduledMinutes
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetExecutedTime(sens []string,
	gs *systemdata.GroundSystem) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		msn.Decrypt()
		msnExp := strings.ToLower(msn.MissionData.Exploitation)
		msnComm := strings.ToLower(msn.MissionData.Communications)
		if gs != nil {
			for _, exp := range gs.Exploitations {
				gsExp := strings.ToLower(exp.Exploitation)
				gsComm := strings.ToLower(exp.CommunicationID)
				if strings.EqualFold(msn.PlatformID, exp.PlatformID) &&
					strings.Contains(gsExp, msnExp) &&
					strings.Contains(gsComm, msnComm) {
					for _, mSen := range msn.MissionData.Sensors {
						if strings.EqualFold(exp.SensorType, mSen.SensorID) &&
							senMax < mSen.ExecutedMinutes {
							senMax = mSen.ExecutedMinutes
						}
					}
				}
			}
		} else {
			for _, sen := range msn.MissionData.Sensors {
				for _, lsen := range sens {
					if strings.EqualFold(sen.SensorID, lsen) &&
						senMax < sen.ExecutedMinutes {
						senMax = sen.ExecutedMinutes
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetAdditional(sens []string,
	gs *systemdata.GroundSystem) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		msn.Decrypt()
		msnExp := strings.ToLower(msn.MissionData.Exploitation)
		msnComm := strings.ToLower(msn.MissionData.Communications)
		if gs != nil {
			for _, exp := range gs.Exploitations {
				gsExp := strings.ToLower(exp.Exploitation)
				gsComm := strings.ToLower(exp.CommunicationID)
				if strings.EqualFold(msn.PlatformID, exp.PlatformID) &&
					strings.Contains(gsExp, msnExp) &&
					strings.Contains(gsComm, msnComm) {
					for _, mSen := range msn.MissionData.Sensors {
						if strings.EqualFold(exp.SensorType, mSen.SensorID) &&
							senMax < mSen.AdditionalMinutes {
							senMax = mSen.AdditionalMinutes
						}
					}
				}
			}
		} else {
			for _, sen := range msn.MissionData.Sensors {
				for _, lsen := range sens {
					if strings.EqualFold(sen.SensorID, lsen) &&
						senMax < sen.AdditionalMinutes {
						senMax = sen.AdditionalMinutes
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetOverlap() uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		msn.Decrypt()
		answer += msn.MissionData.MissionOverlap
	}
	return answer
}

func (mt *MissionType) GetSensorList() []string {
	var answer []string
	for _, msn := range mt.Missions {
		msn.Decrypt()
		for _, sen := range msn.MissionData.Sensors {
			found := false
			for _, s := range answer {
				if strings.EqualFold(sen.SensorID, s) {
					found = true
				}
			}
			if !found {
				answer = append(answer, sen.SensorID)
			}
		}
	}
	return answer
}

type OutageDay struct {
	OutageDate time.Time
	Outages    []metrics.GroundOutage
}
