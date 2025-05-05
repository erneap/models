package sites

import (
	"github.com/erneap/go-models/employees"
	"github.com/erneap/go-models/labor"
)

type Site struct {
	ID              string               `json:"id" bson:"id"`
	Name            string               `json:"name" bson:"name"`
	UtcOffset       float64              `json:"utcOffset" bson:"utcOffset"`
	ShowMids        bool                 `json:"showMids" bson:"showMids"`
	Workcenters     []Workcenter         `json:"workcenters,omitempty" bson:"workcenters,omitempty"`
	LaborCodes      []labor.LaborCode    `json:"laborCodes,omitempty" bson:"laborCodes,omitempty"`
	ForecastReports []ForecastReport     `json:"forecasts,omitempty" bson:"forecasts,omitempty"`
	CofSReports     []CofSReport         `json:"cofs,omitempty" bson:"cofs,omitempty"`
	Employees       []employees.Employee `json:"employees,omitempty" bson:"-"`
}

type BySites []Site

func (c BySites) Len() int { return len(c) }
func (c BySites) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c BySites) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
