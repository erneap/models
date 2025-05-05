package sites

import (
	"sort"
	"time"

	"github.com/erneap/models/labor"
)

type ForecastPeriod struct {
	Month   time.Time   `json:"month" bson:"month"`
	Periods []time.Time `json:"periods,omitempty" bson:"periods,omitempty"`
}

type ByForecastPeriod []ForecastPeriod

func (c ByForecastPeriod) Len() int { return len(c) }
func (c ByForecastPeriod) Less(i, j int) bool {
	return c[i].Month.Before(c[j].Month)
}
func (c ByForecastPeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type ByDate []time.Time

func (c ByDate) Len() int { return len(c) }
func (c ByDate) Less(i, j int) bool {
	return c[i].Before(c[j])
}
func (c ByDate) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type ForecastReport struct {
	ID          int               `json:"id" bson:"id"`
	Name        string            `json:"name" bson:"name"`
	StartDate   time.Time         `json:"startDate" bson:"startDate"`
	EndDate     time.Time         `json:"endDate" bson:"endDate"`
	Periods     []ForecastPeriod  `json:"periods,omitempty" bson:"periods,omitempty"`
	LaborCodes  []labor.LaborCode `json:"laborCodes,omitempty" bson:"laborCodes,omitempty"`
	CompanyID   string            `json:"companyid,omitempty" bson:"companyid,omitempty"`
	SortByFirst bool              `json:"sortfirst,omitempty" bson:"sortfirst,omitempty"`
}

type ByForecastReport []ForecastReport

func (c ByForecastReport) Len() int { return len(c) }
func (c ByForecastReport) Less(i, j int) bool {
	if c[i].StartDate.Equal(c[j].StartDate) {
		if c[i].EndDate.Equal(c[j].EndDate) {
			return c[i].Name < c[j].Name
		}
		return c[i].EndDate.Before(c[j].EndDate)
	}
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByForecastReport) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (r *ForecastReport) ChangePeriodsStart(weekday int) {
	end := time.Date(r.EndDate.Year(), r.EndDate.Month(), r.EndDate.Day(), 0, 0,
		0, 0, time.UTC)
	for int(end.Weekday()) != weekday {
		end = end.AddDate(0, 0, 1)
	}
	end = end.AddDate(0, 0, 1)
	start := time.Date(r.StartDate.Year(), r.StartDate.Month(), r.StartDate.Day(),
		0, 0, 0, 0, time.UTC)
	for int(start.Weekday()) != weekday {
		start = start.AddDate(0, 0, 1)
	}

	// clear the monthly periods of the forecast
	for i, prds := range r.Periods {
		prds.Periods = prds.Periods[:0]
		r.Periods[i] = prds
	}

	for start.Before(end) {
		found := false
		for p, prd := range r.Periods {
			if prd.Month.Year() == start.Year() && prd.Month.Month() == start.Month() {
				prd.Periods = append(prd.Periods, start)
				found = true
				r.Periods[p] = prd
			}
		}
		if !found {
			month := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
			prd := ForecastPeriod{
				Month: month,
			}
			prd.Periods = append(prd.Periods, start)
			r.Periods = append(r.Periods, prd)
		}
		start = start.AddDate(0, 0, 7)
	}
	r.removeUnusedPeriods()
}

func (r *ForecastReport) MovePeriodBetweenMonths(from, to time.Time) {
	var fromPrd ForecastPeriod
	var toPrd ForecastPeriod
	fromPos := -1
	toPos := -1
	sort.Sort(ByForecastPeriod(r.Periods))
	for i, prd := range r.Periods {
		if prd.Month.Equal(from) {
			fromPos = i
			fromPrd = prd
		} else if prd.Month.Equal(to) {
			toPos = i
			toPrd = prd
		}
	}
	if toPos < 0 {
		toPrd = ForecastPeriod{
			Month: to,
		}
		r.Periods = append(r.Periods, toPrd)
		toPos = len(r.Periods) - 1
	}

	sort.Sort(ByDate(fromPrd.Periods))
	if from.Before(to) {
		dPrd := fromPrd.Periods[len(fromPrd.Periods)-1]
		fromPrd.Periods = fromPrd.Periods[:len(fromPrd.Periods)-1]
		toPrd.Periods = append(toPrd.Periods, dPrd)
	} else {
		dPrd := fromPrd.Periods[0]
		fromPrd.Periods = fromPrd.Periods[1:]
		toPrd.Periods = append(toPrd.Periods, dPrd)
	}
	sort.Sort(ByDate(toPrd.Periods))
	r.Periods[fromPos] = fromPrd
	r.Periods[toPos] = toPrd
	sort.Sort(ByForecastPeriod(r.Periods))
	r.removeUnusedPeriods()
}

func (r *ForecastReport) removeUnusedPeriods() {
	for p := len(r.Periods) - 1; p >= 0; p-- {
		if len(r.Periods[p].Periods) == 0 {
			r.Periods = append(r.Periods[:p], r.Periods[p+1:]...)
		}
	}
}

func (r *ForecastReport) AddOutCyclePeriod(dt time.Time) {
	perid := time.Date(dt.Year(), dt.Month(), 1, 0, 0, 0, 0, time.UTC)
	found := false
	for p, prd := range r.Periods {
		if prd.Month.Equal(perid) {
			for _, dtprd := range prd.Periods {
				if dtprd.Equal(dt) {
					found = true
				}
			}
			if !found {
				prd.Periods = append(prd.Periods, dt)
				sort.Sort(ByDate(prd.Periods))
				r.Periods[p] = prd
			}
			found = true
		}
	}
	if !found {
		prd := ForecastPeriod{
			Month: perid,
		}
		prd.Periods = append(prd.Periods, dt)
		r.Periods = append(r.Periods, prd)
		sort.Sort(ByForecastPeriod(r.Periods))
	}
}
