package labor

import (
	"strings"
	"time"
)

type LaborCode struct {
	ChargeNumber     string    `json:"chargeNumber" bson:"chargeNumber"`
	Extension        string    `json:"extension" bson:"extension"`
	CLIN             string    `json:"clin,omitempty" bson:"clin,omitempty"`
	SLIN             string    `json:"slin,omitempty" bson:"slin,omitempty"`
	Location         string    `json:"location,omitempty" bson:"location,omitempty"`
	WBS              string    `json:"wbs,omitempty" bson:"wbs,omitempty"`
	MinimumEmployees int       `json:"minimumEmployees,omitempty" bson:"minimumEmployees,omitempty"`
	NotAssignedName  string    `json:"notAssignedName,omitempty" bson:"notAssignedName,omitempty"`
	HoursPerEmployee float64   `json:"hoursPerEmployee,omitempty" bson:"hoursPerEmployee,omitempty"`
	Exercise         bool      `json:"exercise,omitempty" bson:"exercise,omitempty"`
	StartDate        time.Time `json:"startDate,omitempty" bson:"startDate,omitempty"`
	EndDate          time.Time `json:"endDate,omitempty" bson:"endDate,omitempty"`
}

type ByLaborCode []LaborCode

func (c ByLaborCode) Len() int { return len(c) }
func (c ByLaborCode) Less(i, j int) bool {
	if strings.EqualFold(c[i].ChargeNumber, c[j].ChargeNumber) {
		return c[i].Extension < c[j].Extension
	}
	return c[i].ChargeNumber < c[j].Extension
}
func (c ByLaborCode) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
