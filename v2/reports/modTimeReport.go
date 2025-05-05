package reports

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/erneap/models/v2/employees"
	"github.com/erneap/models/v2/svcs"
	"github.com/xuri/excelize/v2"
)

type WeekPeriod struct {
	Start time.Time
	End   time.Time
}

type ByWeekPeriod []WeekPeriod

func (c ByWeekPeriod) Len() int { return len(c) }
func (c ByWeekPeriod) Less(i, j int) bool {
	return c[i].Start.Before(c[j].Start)
}
func (c ByWeekPeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type MonthPeriod struct {
	Month time.Time
	Weeks []WeekPeriod
}

type ByMonthPeriod []MonthPeriod

func (c ByMonthPeriod) Len() int { return len(c) }
func (c ByMonthPeriod) Less(i, j int) bool {
	return c[i].Month.Before(c[j].Month)
}
func (c ByMonthPeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (mp *MonthPeriod) Label() string {
	return mp.Month.Format("Jan 06")
}

type ModTimeReport struct {
	Report            *excelize.File
	Date              time.Time
	TeamID            string
	SiteID            string
	CompanyID         string
	Styles            map[string]int
	ConditionalStyles map[string]int
	Employees         []employees.Employee
	CurrentAsOf       time.Time
	EndWork           time.Time
	Periods           []MonthPeriod
	MinDate           time.Time
	MaxDate           time.Time
}

func (lr *ModTimeReport) Create() error {
	lr.CurrentAsOf = time.Now().UTC()
	lr.EndWork = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	lr.Styles = make(map[string]int)
	lr.ConditionalStyles = make(map[string]int)
	lr.Report = excelize.NewFile()

	// Get list of forecast reports for the team/site
	lr.MinDate = time.Date(lr.Date.Year(), lr.Date.Month(),
		lr.Date.Day(), 0, 0, 0, 0, time.UTC)
	lr.MaxDate = time.Date(lr.Date.Year(), lr.Date.Month(),
		lr.Date.Day(), 0, 0, 0, 0, time.UTC)
	team, err := svcs.GetTeam(lr.TeamID)
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	found := false
	for _, co := range team.Companies {
		if strings.EqualFold(co.ID, lr.CompanyID) {
			for _, mod := range co.ModPeriods {
				if now.Equal(mod.Start) || now.Equal(mod.End) ||
					(mod.Start.Before(now) && mod.End.After(now)) {
					lr.MinDate = mod.Start
					lr.MaxDate = mod.End
					found = true
				}
			}
		}
	}
	if !found {
		return errors.New("no mod time period for company")
	}

	// set the periods based on minDate for fridays until maxDate
	start := time.Date(lr.MinDate.Year(), lr.MinDate.Month(), lr.MinDate.Day(), 0, 0, 0,
		0, time.UTC)
	for start.Weekday() != time.Friday {
		start = start.AddDate(0, 0, 1)
	}
	// now add periods for dates until max date.
	lr.Periods = []MonthPeriod{}
	var period *MonthPeriod
	for !start.After(lr.MaxDate) {
		if period == nil || period.Month.Month() != start.Month() {
			if period != nil {
				lr.Periods = append(lr.Periods, *period)
			}
			period = &MonthPeriod{
				Month: time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC),
			}
		}
		begin := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
			time.UTC)
		for begin.Weekday() != time.Saturday {
			begin = begin.AddDate(0, 0, -1)
		}
		wk := &WeekPeriod{
			Start: begin,
			End: time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
				time.UTC),
		}
		period.Weeks = append(period.Weeks, *wk)
		sort.Sort(ByWeekPeriod(period.Weeks))
		start = start.AddDate(0, 0, 7)
	}
	if period != nil {
		lr.Periods = append(lr.Periods, *period)
	}
	sort.Sort(ByMonthPeriod(lr.Periods))

	// get employees with assignments for the site that are assigned
	// during the mod period

	emps, err := svcs.GetEmployeesForTeam(lr.TeamID)
	if err != nil {
		return err
	}
	for _, emp := range emps {
		if emp.AtSite(lr.SiteID, lr.MinDate, lr.MaxDate) &&
			strings.EqualFold(emp.CompanyInfo.Company, lr.CompanyID) {
			// get work records for the year inclusive of the two
			// dates
			work, _ := svcs.GetEmployeeWork(emp.ID.Hex(),
				uint(lr.MinDate.Year()))
			if work != nil {
				emp.Work = append(emp.Work, work.Work...)
			}
			if lr.MinDate.Year() != lr.MaxDate.Year() {
				year := lr.MinDate.Year()
				for year < lr.MaxDate.Year() {
					year++
					work, _ = svcs.GetEmployeeWork(emp.ID.Hex(),
						uint(year))
					if work != nil {
						emp.Work = append(emp.Work, work.Work...)
					}
				}
			}
			if emp.GetLastWorkday().After(lr.EndWork) {
				lr.EndWork = emp.GetLastWorkday()
			}
			if emp.HasModTime(lr.MinDate, lr.MaxDate) {
				lr.Employees = append(lr.Employees, emp)
			}
		}
		lr.Report.UpdateLinkedValue()
	}

	//////////////////////////////////////////////////////////
	// Report Creation
	//////////////////////////////////////////////////////////
	lr.CreateStyles()

	lr.CreateModTimeReportSheet()

	lr.Report.DeleteSheet("Sheet1")

	return nil
}

func (lr *ModTimeReport) CreateStyles() error {
	numFmt := "##0.0;[Red]-##0.0;-;@"
	sumFmt := "_(* #,##0.0_);_(* (#,##0.0);_(* \"-\"??_);_(@_)"
	pctFmt := "##0.0%;[Red]##0.0%;-;@"
	style, err := lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 18, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["header"] = style
	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 18, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["headerctr"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 14, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["label"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center", TextRotation: 45,
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["slantwhitelbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99ccff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center", TextRotation: 45,
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["slantbluelbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99cc00"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center", TextRotation: 45,
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["slantgreenlbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center",
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["whitelbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99ccff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center",
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["bluelbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99cc00"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12,
			Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center",
			Vertical: "center",
			WrapText: false},
	})
	if err != nil {
		return err
	}
	lr.Styles["greenlbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 14, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["periods"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 18, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["month"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["monthsumlbl"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["dates"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"da9694"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["monthdate"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffff00"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weeks"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "ffffff", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["peopleheader"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["peoplectr"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["peopleleft"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"fcd5b4"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["liaisonctr"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"fcd5b4"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["liaisonleft"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DA9694"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["sum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DA9694"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &sumFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["monthsum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"bfbfbf"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["actual"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"bfbfbf"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 12, Color: "0000ff", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["forecast"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"bfbfbf"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["cellsum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"c0c0c0"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["graytext"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["whitetext"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99ccff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["bluenum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99ccff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &pctFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["bluepct"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ccffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ltbluenum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ccffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &pctFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ltbluepct"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99cc00"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["greennum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"99cc00"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &pctFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["greenpct"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ccffcc"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ltgreennum"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ccffcc"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &pctFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ltgreenpct"] = style

	// conditional styles
	// slin/wbs yellowed or greened
	format, err := lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"c6efce"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "006100", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
				WrapText: true},
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["greened"] = format

	format, err = lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"ffeb9c"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "9c6500", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
				WrapText: true},
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["yellowed"] = format

	format, err = lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"c6efce"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "9c6500", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center",
				WrapText: true},
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["greenedright"] = format

	format, err = lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"ffeb9c"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "9c6500", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
				WrapText: true},
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["yellowedleft"] = format

	format, err = lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"ffb5b5"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
				WrapText: true},
			CustomNumFmt: &numFmt,
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["cellpink"] = format

	format, err = lr.Report.NewConditionalStyle(
		&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{"ffb5b5"}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 12, Color: "000000", Family: "Calibri Light"},
			Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center",
				WrapText: true},
			CustomNumFmt: &sumFmt,
		},
	)
	if err != nil {
		return err
	}
	lr.ConditionalStyles["pinkright"] = format

	return nil
}

func (lr *ModTimeReport) CreateModTimeReportSheet() error {
	sheetName := "Mod Time"

	lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)

	// set column widths
	lr.Report.SetColWidth(sheetName, "A", "A", 20.0)
	lr.Report.SetColWidth(sheetName, "B", "B", 9.0)
	columns := 2
	for _, period := range lr.Periods {
		sumCol := GetColumn(columns)
		endCol := GetColumn(columns + len(period.Weeks))
		lr.Report.SetColWidth(sheetName, sumCol, endCol, 11.00)
		columns += len(period.Weeks) + 1
	}
	lr.Report.SetColWidth(sheetName, GetColumn(columns), GetColumn(columns), 15.43)

	// headers for page
	style := lr.Styles["label"]
	lr.Report.SetCellStyle(sheetName, "A1", "A2", style)
	lr.Report.MergeCell(sheetName, "A1", "A2")
	lr.Report.SetCellValue(sheetName, "A1", "Name")

	lr.Report.SetCellStyle(sheetName, "B1", "B2", style)
	lr.Report.MergeCell(sheetName, "B1", "B2")
	lr.Report.SetCellValue(sheetName, "B1", "Balance")

	column := 2
	for _, period := range lr.Periods {
		style = lr.Styles["month"]
		lr.Report.SetCellStyle(sheetName, GetCellID(column, 1),
			GetCellID(column+len(period.Weeks), 1), style)
		lr.Report.MergeCell(sheetName, GetCellID(column, 1),
			GetCellID(column+len(period.Weeks), 1))
		lr.Report.SetCellValue(sheetName, GetCellID(column, 1),
			period.Label())
		style = lr.Styles["monthsumlbl"]
		lr.Report.SetCellStyle(sheetName, GetCellID(column, 2),
			GetCellID(column, 3), style)
		lr.Report.SetCellValue(sheetName, GetCellID(column, 2),
			"Month Total")
		lr.Report.SetColOutlineLevel(sheetName, GetColumn(column), 0)
		style = lr.Styles["dates"]
		for i, prd := range period.Weeks {
			cellID := GetCellID(column+i+1, 2)
			lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
			lr.Report.SetCellValue(sheetName, cellID, prd.End.Format("02-Jan"))
			lr.Report.SetColOutlineLevel(sheetName, GetColumn(column+i+1), 1)
		}
		column += len(period.Weeks) + 1
	}

	row := 2
	for _, emp := range lr.Employees {
		row++
		style = lr.Styles["peoplectr"]
		lr.Report.SetCellStyle(sheetName, GetCellID(0, row), GetCellID(1, row),
			style)
		lr.Report.SetCellValue(sheetName, GetCellID(0, row),
			emp.Name.GetLastFirst())
		lr.Report.SetCellValue(sheetName, GetCellID(1, row),
			emp.GetModTime(lr.MinDate, lr.MaxDate))
		column = 1
		var sumlist = []string{}
		for _, period := range lr.Periods {
			column++
			style = lr.Styles["sum"]
			formula := ""
			if len(period.Weeks) > 1 {
				formula = "=SUM(" + GetCellID(column+1, row) + ":" +
					GetCellID(column+(len(period.Weeks)), row) + ")"
			} else {
				formula = "=" + GetCellID(column+1, row)
			}
			cellID := GetCellID(column, row)
			sumlist = append(sumlist, cellID)
			lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
			lr.Report.SetCellFormula(sheetName, cellID, formula)
			for _, prd := range period.Weeks {
				column++
				cellID = GetCellID(column, row)
				style = lr.Styles["actual"]
				lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
				format := lr.ConditionalStyles["cellpink"]
				lr.Report.SetConditionalFormat(sheetName, cellID,
					[]excelize.ConditionalFormatOptions{
						{Type: "cell", Criteria: "==", Format: format, Value: "0"},
					})
				lr.Report.SetCellValue(sheetName, cellID, emp.GetModTime(prd.Start,
					prd.End))
			}
		}
	}
	return nil
}
