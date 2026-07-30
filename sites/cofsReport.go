package sites

import (
	"time"

	"github.com/erneap/models/labor"
)

type CofSCompany struct {
	ID             string            `json:"id" bson:"id"`
	SignatureBlock string            `json:"signature" bson:"signature"`
	LaborCodes     []labor.LaborCode `json:"laborcodes,omitempty" bson:"laborcodes,omitempty"`
	SortID         int               `json:"sortid" bson:"sortid"`
	AddExercises   bool              `json:"exercises" bson:"exercises"`
}

type ByCofSCompany []CofSCompany

func (c ByCofSCompany) Len() int { return len(c) }
func (c ByCofSCompany) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByCofSCompany) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type CofSSection struct {
	ID             int               `json:"id" bson:"id"`
	CompanyID      string            `json:"company" bson:"company"`
	Label          string            `json:"label" bson:"label"`
	SignatureBlock string            `json:"signature" bson:"signature"`
	LaborCodes     []labor.LaborCode `json:"laborcodes,omitempty" bson:"laborcodes,omitempty"`
	ShowUnit       bool              `json:"showunit" bson:"showunit"`
}

type ByCofSSection []CofSSection

func (c ByCofSSection) Len() int { return len(c) }
func (c ByCofSSection) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByCofSSection) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type CofSReport struct {
	ID             int           `json:"id" bson:"id"`
	Name           string        `json:"name" bson:"name"`
	ShortName      string        `json:"shortname" bson:"shortname"`
	AssociatedUnit string        `json:"unit" bson:"unit"`
	StartDate      time.Time     `json:"startdate" bson:"startdate"`
	EndDate        time.Time     `json:"enddate" bson:"enddate"`
	Companies      []CofSCompany `json:"companies,omitempty" bson:"companies,omitempty"`
	Sections       []CofSSection `json:"sections,omitempty" bson:"sections,omitempty"`
}

type ByCofSReport []CofSReport

func (c ByCofSReport) Len() int { return len(c) }
func (c ByCofSReport) Less(i, j int) bool {
	if c[i].StartDate.Equal(c[j].StartDate) {
		return c[i].Name < c[j].Name
	}
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByCofSReport) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
