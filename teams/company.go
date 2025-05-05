package teams

import (
	"sort"
	"strings"
	"time"
)

type CompanyHoliday struct {
	ID          string      `json:"id" bson:"id"`
	Name        string      `json:"name" bson:"name"`
	SortID      uint        `json:"sort" bson:"sort"`
	ActualDates []time.Time `json:"actualdates,omitempty" bson:"actualdates,omitempty"`
}

type ByCompanyHoliday []CompanyHoliday

func (c ByCompanyHoliday) Len() int { return len(c) }
func (c ByCompanyHoliday) Less(i, j int) bool {
	if c[i].ID == c[j].ID {
		return c[i].SortID < c[j].SortID
	}
	return strings.EqualFold(c[i].ID, "H")
}
func (c ByCompanyHoliday) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (ch *CompanyHoliday) GetActual(year int) *time.Time {
	start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(year+1, 1, 1, 0, 0, 0, 0, time.UTC)
	for _, actual := range ch.ActualDates {
		if (actual.Equal(start) || actual.After(start)) && actual.Before(end) {
			return &actual
		}
	}
	return nil
}

func (ch *CompanyHoliday) Purge(date time.Time) {
	for i := len(ch.ActualDates) - 1; i >= 0; i-- {
		if ch.ActualDates[i].Before(date) {
			ch.ActualDates = append(ch.ActualDates[:i], ch.ActualDates[i+1:]...)
		}
	}
}

type Company struct {
	ID             string           `json:"id" bson:"id"`
	Name           string           `json:"name" bson:"name"`
	IngestType     string           `json:"ingest" bson:"ingest"`
	IngestPeriod   int              `json:"ingestPeriod,omitempty" bson:"ingestPeriod,omitempty"`
	IngestStartDay int              `json:"startDay,omitempty" bson:"startDay,omitempty"`
	IngestPwd      string           `json:"ingestPwd" bson:"ingestPwd"`
	Holidays       []CompanyHoliday `json:"holidays,omitempty" bson:"holidays,omitempty"`
	ModPeriods     []ModPeriod      `json:"modperiods,omitempty" bson:"modperiods,omitempty"`
}

type ByCompany []Company

func (c ByCompany) Len() int { return len(c) }
func (c ByCompany) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c ByCompany) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c *Company) Purge(date time.Time) {
	for h, hol := range c.Holidays {
		hol.Purge(date)
		c.Holidays[h] = hol
	}
}

func (c *Company) HasModPeriod(date time.Time) bool {
	for _, mod := range c.ModPeriods {
		if date.Equal(mod.Start) || date.Equal(mod.End) ||
			(date.After(mod.Start) && date.Before(mod.End)) {
			return true
		}
	}
	return false
}

func (c *Company) AddModPeriod(year int, start, end time.Time) {
	for m, mod := range c.ModPeriods {
		if mod.Year == year {
			mod.Start = start
			mod.End = end
			c.ModPeriods[m] = mod
			return
		}
	}
	mod := ModPeriod{
		Year:  year,
		Start: start,
		End:   end,
	}
	c.ModPeriods = append(c.ModPeriods, mod)
	sort.Sort(ByModPeriod(c.ModPeriods))
}

func (c *Company) UpdateModPeriod(year int, field string, date time.Time) {
	for m, mod := range c.ModPeriods {
		if mod.Year == year {
			if strings.ToLower(field) == "start" {
				mod.Start = date
			} else if strings.ToLower(field) == "end" {
				mod.End = date
			}
			c.ModPeriods[m] = mod
		}
	}
}

func (c *Company) DeleteModPeriod(year int) {
	pos := -1
	for m, mod := range c.ModPeriods {
		if mod.Year == year {
			pos = m
		}
	}
	if pos >= 0 {
		c.ModPeriods = append(c.ModPeriods[:pos], c.ModPeriods[pos+1:]...)
	}
}
