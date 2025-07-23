package plans

import (
	"sort"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Reading struct {
	Id         uint   `bson:"id" json:"id"`
	Book       string `bson:"book" json:"book"`
	Chapter    uint   `bson:"chapter" json:"chapter"`
	VerseStart uint   `bson:"verseStart" json:"verseStart"`
	VerseEnd   uint   `bson:"verseEnd" json:"verseEnd"`
}

type ByReadings []Reading

func (c ByReadings) Len() int { return len(c) }
func (c ByReadings) Less(i, j int) bool {
	return c[i].Id < c[j].Id
}
func (c ByReadings) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type PlanDay struct {
	DayOfMonth uint      `bson:"dayOfMonth" json:"dayOfMonth"`
	Readings   []Reading `bson:"readings,omitempty" json:"readings,omitempty"`
}

type ByPlanDays []PlanDay

func (c ByPlanDays) Len() int { return len(c) }
func (c ByPlanDays) Less(i, j int) bool {
	return c[i].DayOfMonth < c[j].DayOfMonth
}
func (c ByPlanDays) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (pd *PlanDay) AddReading(r Reading) {
	found := false
	for i, read := range pd.Readings {
		if read.Id == r.Id {
			found = true
			pd.Readings[i] = r
		}
	}
	if !found {
		pd.Readings = append(pd.Readings, r)
		sort.Sort(ByReadings(pd.Readings))
	}
}

func (pd *PlanDay) UpdateReading(id uint, field, value string) {
	for i, r := range pd.Readings {
		if r.Id == id {
			switch strings.ToLower(field) {
			case "book":
				r.Book = value
			case "chapter":
				uiValue, err := strconv.ParseUint(value, 10, 64)
				if err == nil {
					r.Chapter = uint(uiValue)
				}
			case "versestart":
				uiValue, err := strconv.ParseUint(value, 10, 64)
				if err == nil {
					r.VerseStart = uint(uiValue)
				}
			case "verseend":
				uiValue, err := strconv.ParseUint(value, 10, 64)
				if err == nil {
					r.VerseEnd = uint(uiValue)
				}
			}
			pd.Readings[i] = r
		}
	}
}

func (pd *PlanDay) DeleteReading(id uint) {
	found := -1
	for i, r := range pd.Readings {
		if r.Id == id {
			found = i
		}
	}
	if found >= 0 {
		pd.Readings = append(pd.Readings[:found], pd.Readings[found+1:]...)
	}
}

type PlanMonth struct {
	Month uint      `bson:"month" json:"month"`
	Days  []PlanDay `bson:"days,omitempty" json:"days,omitempty"`
}

type ByPlanMonth []PlanMonth

func (c ByPlanMonth) Len() int { return len(c) }
func (c ByPlanMonth) Less(i, j int) bool {
	return c[i].Month < c[j].Month
}
func (c ByPlanMonth) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (pm *PlanMonth) SetDays(num int) {
	if len(pm.Days) < num {
		for n := len(pm.Days); n < num; n++ {
			found := false
			for _, d := range pm.Days {
				if d.DayOfMonth == uint(n) {
					found = true
				}
			}
			if !found {
				day := PlanDay{
					DayOfMonth: uint(n),
				}
				pm.Days = append(pm.Days, day)
			}
		}
		sort.Sort(ByPlanDays(pm.Days))
	} else if len(pm.Days) > num {
		sort.Sort(ByPlanDays(pm.Days))
		pm.Days = pm.Days[:num]
	}
}

type Plan struct {
	ID     primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name   string             `bson:"name" json:"name"`
	Months []PlanMonth        `bson:"months,omitempty" json:"months,omitempty"`
}

type ByPlans []Plan

func (c ByPlans) Len() int { return len(c) }
func (c ByPlans) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c ByPlans) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p *Plan) SetMonths(num int) {
	if len(p.Months) < num {
		for n := len(p.Months); n < num; n++ {
			found := false
			for _, m := range p.Months {
				if m.Month == uint(n) {
					found = true
				}
			}
			if !found {
				month := PlanMonth{
					Month: uint(n),
				}
				p.Months = append(p.Months, month)
			}
		}
		sort.Sort(ByPlanMonth(p.Months))

	} else if len(p.Months) > num {
		sort.Sort(ByPlanMonth(p.Months))
		p.Months = p.Months[:num]
	}
}
