package reports

import (
	"strings"
	"time"

	"github.com/erneap/models/v2/metrics"
	"github.com/erneap/models/v2/systemdata"
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
				for _, sen := range msn.Sensors {
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
				if !msn.Aborted && !msn.Cancelled &&
					!msn.IndefDelay {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				if !msn.Aborted && !msn.Cancelled &&
					!msn.IndefDelay {
					for _, sen := range msn.Sensors {
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
				if msn.Cancelled || msn.IndefDelay {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				if msn.Cancelled || msn.IndefDelay {
					for _, sen := range msn.Sensors {
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
				if msn.Aborted {
					answer++
				}
			}
		} else {
			for _, msn := range mt.Missions {
				found := false
				if msn.Aborted {
					for _, sen := range msn.Sensors {
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
	gs *systemdata.GroundSystem, enclave string) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		if gs.CheckForUse && msn.EquipmentInUse(gs.ID) {
			if gs != nil {
				for _, mSen := range msn.Sensors {
					if gs.UseMissionSensor(msn.PlatformID, mSen.SensorID, msn.Exploitation,
						msn.Communications, enclave) {
						if senMax < mSen.PreflightMinutes {
							senMax = mSen.PreflightMinutes
						}
					}
				}
			} else {
				for _, sen := range msn.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) &&
							senMax < sen.PreflightMinutes {
							senMax = sen.PreflightMinutes
						}
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetPostmissionTime(sens []string,
	gs *systemdata.GroundSystem, enclave string) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		if gs.CheckForUse && msn.EquipmentInUse(gs.ID) {
			if gs != nil {
				for _, mSen := range msn.Sensors {
					if gs.UseMissionSensor(msn.PlatformID, mSen.SensorID, msn.Exploitation,
						msn.Communications, enclave) {
						if senMax < mSen.PostflightMinutes {
							senMax = mSen.PostflightMinutes
						}
					}
				}
			} else {
				for _, sen := range msn.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) &&
							senMax < sen.PostflightMinutes {
							senMax = sen.PostflightMinutes
						}
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetScheduledTime(sens []string,
	gs *systemdata.GroundSystem, enclave string) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		if gs.CheckForUse && msn.EquipmentInUse(gs.ID) {
			if gs != nil {
				for _, mSen := range msn.Sensors {
					if gs.UseMissionSensor(msn.PlatformID, mSen.SensorID, msn.Exploitation,
						msn.Communications, enclave) {
						if senMax < mSen.ScheduledMinutes {
							senMax = mSen.ScheduledMinutes
						}
					}
				}
			} else {
				for _, sen := range msn.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) &&
							senMax < sen.ScheduledMinutes {
							senMax = sen.ScheduledMinutes
						}
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetExecutedTime(sens []string,
	gs *systemdata.GroundSystem, enclave string) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		if gs.CheckForUse && msn.EquipmentInUse(gs.ID) {
			if gs != nil {
				for _, mSen := range msn.Sensors {
					if gs.UseMissionSensor(msn.PlatformID, mSen.SensorID, msn.Exploitation,
						msn.Communications, enclave) {
						if senMax < mSen.ExecutedMinutes {
							senMax = mSen.ExecutedMinutes
						}
					}
				}
			} else {
				for _, sen := range msn.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) &&
							senMax < sen.ExecutedMinutes {
							senMax = sen.ExecutedMinutes
						}
					}
				}
			}
		}
		answer += senMax
	}
	return answer
}

func (mt *MissionType) GetAdditional(sens []string,
	gs *systemdata.GroundSystem, enclave string) uint {
	answer := uint(0)
	for _, msn := range mt.Missions {
		senMax := uint(0)
		if gs.CheckForUse && msn.EquipmentInUse(gs.ID) {
			if gs != nil {
				for _, mSen := range msn.Sensors {
					if gs.UseMissionSensor(msn.PlatformID, mSen.SensorID, msn.Exploitation,
						msn.Communications, enclave) {
						if senMax < mSen.AdditionalMinutes {
							senMax = mSen.AdditionalMinutes
						}
					}
				}
			} else {
				for _, sen := range msn.Sensors {
					for _, lsen := range sens {
						if strings.EqualFold(sen.SensorID, lsen) &&
							senMax < sen.AdditionalMinutes {
							senMax = sen.AdditionalMinutes
						}
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
		answer += msn.MissionOverlap
	}
	return answer
}

func (mt *MissionType) GetSensorList() []string {
	var answer []string
	for _, msn := range mt.Missions {
		for _, sen := range msn.Sensors {
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
