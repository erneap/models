package reports

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/models/employees"
	"github.com/erneap/models/labor"
	"github.com/erneap/models/svcs"
	"github.com/erneap/models/teams"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/maps"
)

type LeaveMonth struct {
	Month   *time.Time
	Holiday *teams.CompanyHoliday
	Disable bool
	Periods []LeavePeriod
}

type ByLeaveMonth []LeaveMonth

func (c ByLeaveMonth) Len() int { return len(c) }
func (c ByLeaveMonth) Less(i, j int) bool {
	if c[i].Holiday != nil {
		if c[i].Holiday.ID == c[j].Holiday.ID {
			return c[i].Holiday.SortID < c[j].Holiday.SortID
		}
		return strings.EqualFold(c[i].Holiday.ID, "H")
	}
	return c[i].Month.Before(*c[j].Month)
}
func (c ByLeaveMonth) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (lm *LeaveMonth) GetPTOActual() float64 {
	hours := 0.0
	for _, lvPer := range lm.Periods {
		if strings.EqualFold(lvPer.Code, "V") {
			for _, lv := range lvPer.Leaves {
				if strings.EqualFold(lv.Status, "actual") {
					hours += lv.Hours
				}
			}
		}
	}
	return hours
}

func (lm *LeaveMonth) GetPTOSchedule() float64 {
	hours := 0.0
	for _, lvPer := range lm.Periods {
		if strings.EqualFold(lvPer.Code, "V") {
			for _, lv := range lvPer.Leaves {
				if !strings.EqualFold(lv.Status, "actual") {
					hours += lv.Hours
				}
			}
		}
	}
	return hours
}

func (lm *LeaveMonth) GetHours() float64 {
	hours := 0.0
	for _, lvPer := range lm.Periods {
		for _, lv := range lvPer.Leaves {
			hours += lv.Hours
		}
	}
	return hours
}

func (lm *LeaveMonth) GetHolidayHours() float64 {
	hours := 0.0
	for _, lvPer := range lm.Periods {
		if strings.EqualFold(lvPer.Code, "H") {
			for _, lv := range lvPer.Leaves {
				if strings.EqualFold(lv.Status, "actual") {
					hours += lv.Hours
				}
			}
		}
	}
	return hours
}

type LeavePeriod struct {
	Code      string
	StartDate time.Time
	EndDate   time.Time
	Status    string
	Leaves    []employees.LeaveDay
}

func (lp *LeavePeriod) GetHours() float64 {
	hours := 0.0
	for _, lv := range lp.Leaves {
		hours += lv.Hours
	}
	return hours
}

type ByLeavePeriod []LeavePeriod

func (c ByLeavePeriod) Len() int { return len(c) }
func (c ByLeavePeriod) Less(i, j int) bool {
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByLeavePeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type LeaveReport struct {
	Report    *excelize.File
	Year      int
	TeamID    string
	SiteID    string
	CompanyID string
	BHolidays bool
	Holidays  []LeaveMonth
	Workcodes map[string]labor.Workcode
	Styles    map[string]int
	Employees []employees.Employee
	Offset    float64
}

func (lr *LeaveReport) Create() error {
	lr.Styles = make(map[string]int)
	lr.Workcodes = make(map[string]labor.Workcode)
	lr.Report = excelize.NewFile()

	// get employees with assignments for the site that are assigned
	// during the year.
	startDate := time.Date(lr.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(lr.Year, 12, 31, 23, 59, 59, 0, time.UTC)
	emps, err := svcs.GetEmployeesForTeam(lr.TeamID)
	if err != nil {
		return err
	}

	for _, emp := range emps {
		if emp.AtSite(lr.SiteID, startDate, endDate) {
			if strings.EqualFold(emp.CompanyInfo.Company, lr.CompanyID) {
				lr.Employees = append(lr.Employees, emp)
			}
		}
	}

	sort.Sort(employees.ByEmployees(lr.Employees))

	team, err := svcs.GetTeam(lr.TeamID)
	if err != nil {
		return err
	}
	for _, com := range team.Companies {
		if strings.EqualFold(com.ID, lr.CompanyID) {
			lr.BHolidays = len(com.Holidays) > 0
			for _, hol := range com.Holidays {
				holiday := &teams.CompanyHoliday{
					ID:     hol.ID,
					SortID: hol.SortID,
					Name:   hol.Name,
				}
				holiday.ActualDates = append(holiday.ActualDates, hol.ActualDates...)
				h := LeaveMonth{
					Holiday: holiday,
				}
				lr.Holidays = append(lr.Holidays, h)
			}
			sort.Sort(ByLeaveMonth(lr.Holidays))
		}
	}
	for _, wc := range team.Workcodes {
		if wc.IsLeave {
			lr.Workcodes[wc.Id] = wc
		}
	}
	for _, site := range team.Sites {
		if strings.EqualFold(site.ID, lr.SiteID) {
			lr.Offset = site.UtcOffset
		}
	}

	lr.CreateStyles()

	lr.CreateLeaveListing()

	lr.CreateFullMonthlyReference()

	lr.CreateMinumimMonthlyReference()

	lr.Report.DeleteSheet("Sheet1")

	return nil
}

func (lr *LeaveReport) CreateStyles() error {
	style, err := lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"00ffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["ptoname"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"800000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["currentasof"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"cccccc"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["section"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"cccccc"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 8, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["sectionlbl"] = style

	numFmt := "0.0;mm/dd/yyyy;@"
	intFmt := "0;@"
	balFmt := "0.0;[Red]-0.0;@"
	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["hollblactual"] = style
	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &intFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["holdaysleft"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["hollblsched"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["holdateactual"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["holdatesched"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["holtaken"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "00ffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["holprojected"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"999999"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["disabled"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"999999"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["disabledlt"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "00ffff"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["month"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ptodates"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ptotaken"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "3366ff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["ptorequest"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffff00"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &balFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["balance"] = style

	for k, v := range lr.Workcodes {
		style, err = lr.Report.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{v.BackColor}, Pattern: 1},
			Font: &excelize.Font{Bold: false, Size: 10, Color: v.TextColor},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
				WrapText: true},
			CustomNumFmt: &numFmt,
		})
		if err == nil {
			lr.Styles[k] = style
		}
	}

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["header"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"009933"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["monthup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"009933"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["monthdn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekdayup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekdaydn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekday"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"66ffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekendup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"66ffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekenddn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"66ffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["weekend"] = style

	pto := lr.Workcodes["V"]
	holiday := lr.Workcodes["H"]

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"009933"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["totalup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"009933"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["totaldn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"009933"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "ffffff"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["total"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{holiday.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: holiday.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["holup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{holiday.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: holiday.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["holdn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{holiday.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: holiday.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["hol"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 0},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{pto.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: pto.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["ptoup"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{pto.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: pto.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["ptodn"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{pto.BackColor}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: pto.TextColor},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		CustomNumFmt: &numFmt,
	})
	if err != nil {
		return err
	}
	lr.Styles["pto"] = style

	style, err = lr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 0},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"ffffff"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	lr.Styles["legend"] = style

	return nil
}

func (lr *LeaveReport) CreateLeaveListing() error {
	sheetName := strconv.Itoa(lr.Year) + " PTO-Hol"
	lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)

	extendedWidth := 3
	if lr.BHolidays {
		extendedWidth += 4
	}

	// add current as of reference
	curStyle := lr.Styles["currentasof"]
	lr.Report.SetCellStyle(sheetName, GetCellID(0, 1),
		GetCellID(extendedWidth, 1), curStyle)
	lr.Report.MergeCell(sheetName, GetCellID(0, 1),
		GetCellID(extendedWidth, 1))
	lr.Report.SetCellValue(sheetName, GetCellID(0, 1),
		"Current As Of: "+time.Now().Format("01/02/2006"))

	// Freeze the first row
	lr.Report.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		YSplit:      1,
		TopLeftCell: "A2",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: "A2", ActiveCell: "A2", Pane: "bottomLeft"},
		},
	})

	// months
	var months []LeaveMonth
	for i := 0; i < 12; i++ {
		dtMonth := time.Date(lr.Year, time.Month(i+1), 1, 0, 0, 0, 0, time.UTC)
		month := LeaveMonth{
			Month:   &dtMonth,
			Holiday: nil,
		}
		months = append(months, month)
	}

	// set column widths
	if lr.BHolidays {
		lr.Report.SetColWidth(sheetName, GetColumn(0), GetColumn(0), 4.5)
		lr.Report.SetColWidth(sheetName, GetColumn(1), GetColumn(1), 13.0)
		lr.Report.SetColWidth(sheetName, GetColumn(2), GetColumn(2), 30.0)
		lr.Report.SetColWidth(sheetName, GetColumn(3), GetColumn(3), 7.0)
		lr.Report.SetColWidth(sheetName, GetColumn(4), GetColumn(4), 30.0)
		lr.Report.SetColWidth(sheetName, GetColumn(5), GetColumn(8), 7.0)
	} else {
		lr.Report.SetColWidth(sheetName, GetColumn(0), GetColumn(0), 30.0)
		lr.Report.SetColWidth(sheetName, GetColumn(1), GetColumn(4), 7.0)
	}

	row := 2
	for _, emp := range lr.Employees {
		annual := 0.0
		carry := 0.0
		for _, bal := range emp.Balances {
			if bal.Year == lr.Year {
				annual = bal.Annual
				carry = bal.Carryover
			}
		}
		row++
		// name row for the employee
		style := lr.Styles["ptoname"]
		lr.Report.SetCellStyle(sheetName, GetCellID(0, row),
			GetCellID(extendedWidth, row), style)
		lr.Report.MergeCell(sheetName, GetCellID(0, row),
			GetCellID(extendedWidth, row))
		lr.Report.SetCellValue(sheetName, GetCellID(0, row), emp.Name.GetLastFirst())

		col := 0
		row++
		style = lr.Styles["section"]
		if lr.BHolidays {
			lr.Report.SetCellStyle(sheetName, GetCellID(0, row),
				GetCellID(3, row), style)
			lr.Report.MergeCell(sheetName, GetCellID(0, row),
				GetCellID(3, row))
			lr.Report.SetCellValue(sheetName, GetCellID(0, row), "Holidays")
			col = 4
		}
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col+3, row), style)
		lr.Report.MergeCell(sheetName, GetCellID(col, row),
			GetCellID(col+3, row))
		lr.Report.SetCellValue(sheetName, GetCellID(col, row), "Leaves")

		col = 0
		row++
		style = lr.Styles["sectionlbl"]
		if lr.BHolidays {
			lr.Report.SetCellStyle(sheetName, GetCellID(0, row), GetCellID(3, row),
				style)
			lr.Report.SetCellValue(sheetName, GetCellID(0, row), "")
			lr.Report.SetCellValue(sheetName, GetCellID(1, row), "Reference Date")
			lr.Report.SetCellValue(sheetName, GetCellID(3, row), "Hours")
			lr.Report.SetCellRichText(sheetName, GetCellID(2, row),
				[]excelize.RichTextRun{
					{
						Text: "Date Taken (",
						Font: &excelize.Font{
							Bold:  true,
							Size:  8,
							Color: "000000",
						},
					},
					{
						Text: "Projected",
						Font: &excelize.Font{
							Bold:  true,
							Size:  8,
							Color: "3366ff",
						},
					},
					{
						Text: ")",
						Font: &excelize.Font{
							Bold:  true,
							Size:  8,
							Color: "000000",
						},
					},
				})
			col = 4
		}
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row), GetCellID(col+3, row),
			style)
		lr.Report.SetCellValue(sheetName, GetCellID(col+2, row), "Taken")
		lr.Report.SetCellRichText(sheetName, GetCellID(col+3, row),
			[]excelize.RichTextRun{
				{
					Text: "Request",
					Font: &excelize.Font{
						Bold:  true,
						Size:  8,
						Color: "3366ff",
					},
				},
			})
		lr.Report.MergeCell(sheetName, GetCellID(col, row), GetCellID(col+1, row))
		lr.Report.SetCellRichText(sheetName, GetCellID(col, row),
			[]excelize.RichTextRun{
				{
					Text: "Leave Taken (",
					Font: &excelize.Font{
						Bold:  true,
						Size:  8,
						Color: "000000",
					},
				},
				{
					Text: "Projected",
					Font: &excelize.Font{
						Bold:  true,
						Size:  8,
						Color: "3366ff",
					},
				},
				{
					Text: ")",
					Font: &excelize.Font{
						Bold:  true,
						Size:  8,
						Color: "000000",
					},
				},
			})
		holRow := 0
		lvRow := 0
		startAsgmt := emp.Assignments[0]
		endAsgmt := emp.Assignments[len(emp.Assignments)-1]
		// clear months
		for m, month := range months {
			startMonth := month.Month.AddDate(0, 1, 0)
			month.Periods = month.Periods[:0]
			month.Disable = startAsgmt.StartDate.After(startMonth) ||
				endAsgmt.EndDate.Before(*month.Month)
			months[m] = month
		}
		// clear holidays periods
		for h, hol := range lr.Holidays {
			actual := hol.Holiday.GetActual(lr.Year)
			hol.Periods = hol.Periods[:0]
			if actual != nil {
				hol.Disable = hol.Holiday.ID[0:1] == "H" &&
					startAsgmt.StartDate.After(*actual) ||
					endAsgmt.EndDate.Before(*actual)
			} else {
				hol.Disable = false
			}
			lr.Holidays[h] = hol
		}

		sort.Sort(employees.ByLeaveDay(emp.Leaves))
		std := emp.GetStandardWorkday(time.Date(lr.Year, 1, 1, 0, 0, 0, 0, time.UTC))

		// create employee holiday list and other leave list (array)
		var empHolidays []employees.LeaveDay
		var empOtherLeave []employees.LeaveDay

		for _, lv := range emp.Leaves {
			lv.Used = false
			if lv.LeaveDate.UTC().Year() == lr.Year &&
				!lv.LeaveDate.Before(startAsgmt.StartDate) &&
				!lv.LeaveDate.After(endAsgmt.EndDate) {
				if strings.ToLower(lv.Code) == "h" {
					empHolidays = append(empHolidays, lv)
				} else {
					empOtherLeave = append(empOtherLeave, lv)
				}
			}
		}

		// process employee holidays first,
		// 1.  Place the tagged leaves to the proper holiday
		for h, lv := range empHolidays {
			if lv.TagDay != "" && !lv.Used {
				for c, cHol := range lr.Holidays {
					cHolID := cHol.Holiday.ID + strconv.Itoa(int(cHol.Holiday.SortID))
					if strings.EqualFold(cHolID, lv.TagDay) &&
						!cHol.Disable && cHol.GetHolidayHours()+lv.Hours <= 8.0 {
						prd := LeavePeriod{
							Code:      lv.Code,
							StartDate: lv.LeaveDate,
							EndDate:   lv.LeaveDate,
							Status:    lv.Status,
						}
						prd.Leaves = append(prd.Leaves, lv)
						cHol.Periods = append(cHol.Periods, prd)
						lr.Holidays[c] = cHol
						lv.Used = true
						empHolidays[h] = lv
					}
				}
			}
		}

		// 2.  Actuals displayed first
		for h, lv := range empHolidays {
			for c, cHol := range lr.Holidays {
				if strings.ToLower(cHol.Holiday.ID) == "h" &&
					strings.ToLower(lv.Status) == "actual" &&
					!lv.Used && !cHol.Disable &&
					cHol.GetHolidayHours()+lv.Hours <= 8.0 {
					prd := LeavePeriod{
						Code:      lv.Code,
						StartDate: lv.LeaveDate,
						EndDate:   lv.LeaveDate,
						Status:    lv.Status,
					}
					prd.Leaves = append(prd.Leaves, lv)
					cHol.Periods = append(cHol.Periods, prd)
					lr.Holidays[c] = cHol
					lv.Used = true
					empHolidays[h] = lv
				}
			}
		}

		// 3.  loop through holidays and add leave to period if equal to the
		// reference date
		for c := 0; c < len(lr.Holidays); c++ {
			cHol := lr.Holidays[c]
			start := cHol.Holiday.GetActual(lr.Year)
			if len(cHol.Periods) == 0 && start != nil {
				end := start.AddDate(0, 0, 1)
				bFound := false
				for h := 0; h < len(empHolidays) && !bFound; h++ {
					lv := empHolidays[h]
					if !lv.Used && lv.LeaveDate.Compare(*start) >= 0 &&
						lv.LeaveDate.Compare(end) < 0 && !cHol.Disable {
						prd := LeavePeriod{
							Code:      lv.Code,
							StartDate: lv.LeaveDate,
							EndDate:   lv.LeaveDate,
							Status:    lv.Status,
						}
						prd.Leaves = append(prd.Leaves, lv)
						cHol.Periods = append(cHol.Periods, prd)
						lr.Holidays[c] = cHol
						lv.Used = true
						empHolidays[h] = lv
						bFound = true
					}
				}
			}
		}
		// 4.  put remaining holidays in unused company holidays up to 8 hours per holiday.
		for e, eHol := range empHolidays {
			bFound := eHol.Used
			for c, cHol := range lr.Holidays {
				if !bFound && !cHol.Disable {
					if cHol.GetHours() < 8.0 {
						if cHol.GetHours()+eHol.Hours <= 8.0 {
							bFound = true
							prd := LeavePeriod{
								Code:      eHol.Code,
								StartDate: eHol.LeaveDate,
								EndDate:   eHol.LeaveDate,
								Status:    eHol.Status,
							}
							prd.Leaves = append(prd.Leaves, eHol)
							cHol.Periods = append(cHol.Periods, prd)
							lr.Holidays[c] = cHol
							eHol.Used = true
							empHolidays[e] = eHol
						}
					}
				}
			}
		}

		// 5.  if there are unused holidays, plug into any disabled holidays
		for _, eHol := range empHolidays {
			if !eHol.Used {
				bFound := false
				for c := 0; c < len(lr.Holidays) && !bFound; c++ {
					cHol := lr.Holidays[c]
					if cHol.Disable {
						bFound = true
						prd := LeavePeriod{
							Code:      eHol.Code,
							StartDate: eHol.LeaveDate,
							EndDate:   eHol.LeaveDate,
							Status:    eHol.Status,
						}
						prd.Leaves = append(prd.Leaves, eHol)
						cHol.Periods = append(cHol.Periods, prd)
						lr.Holidays[c] = cHol
					}
				}
			}
		}

		for _, lv := range empOtherLeave {
			for m, month := range months {
				if lv.LeaveDate.Hour() != 0 {
					delta := time.Hour * time.Duration(lr.Offset)
					lv.LeaveDate = lv.LeaveDate.Add(delta)
				}
				if month.Month.Year() == lv.LeaveDate.Year() &&
					month.Month.Month() == lv.LeaveDate.Month() {
					bFound := false
					for p, prd := range month.Periods {
						if strings.EqualFold(prd.Code, lv.Code) &&
							strings.EqualFold(prd.Status, lv.Status) &&
							prd.EndDate.Day()+1 == lv.LeaveDate.Day() &&
							!bFound && lv.Hours >= std && prd.GetHours() >= std {
							bFound = true
							prd.Leaves = append(prd.Leaves, lv)
							prd.EndDate = lv.LeaveDate
							month.Periods[p] = prd
						}
					}
					if !bFound {
						prd := LeavePeriod{
							Code:      lv.Code,
							StartDate: lv.LeaveDate,
							EndDate:   lv.LeaveDate,
							Status:    lv.Status,
						}
						prd.Leaves = append(prd.Leaves, lv)
						month.Periods = append(month.Periods, prd)
					}
					months[m] = month
				}
			}
		}
		now := time.Now().UTC()
		col = 0
		var richText []excelize.RichTextRun
		if lr.BHolidays {
			sort.Sort(ByLeaveMonth(lr.Holidays))
			for _, hol := range lr.Holidays {
				holRow++
				sStyle := "hollblactual"
				if hol.Disable {
					sStyle = "disabled"
				} else {
					if hol.Holiday.GetActual(lr.Year) != nil &&
						hol.Holiday.GetActual(lr.Year).After(now) {
						sStyle = "hollblsched"
					}
				}
				style := lr.Styles[sStyle]
				lr.Report.SetCellStyle(sheetName, GetCellID(0, row+holRow),
					GetCellID(3, row+holRow), style)
				lr.Report.SetCellValue(sheetName, GetCellID(0, row+holRow),
					fmt.Sprintf("%s%d", hol.Holiday.ID, hol.Holiday.SortID))
				if hol.Holiday.GetActual(lr.Year) != nil {
					lr.Report.SetCellValue(sheetName, GetCellID(1, row+holRow),
						hol.Holiday.GetActual(lr.Year).Format("02-Jan-06"))
				} else {
					lr.Report.SetCellValue(sheetName, GetCellID(1, row+holRow), "")
				}
				lr.Report.SetCellValue(sheetName, GetCellID(3, row+holRow),
					hol.GetHolidayHours())
				richText = richText[:0]
				for _, prd := range hol.Periods {
					for _, lv := range prd.Leaves {
						if strings.EqualFold(lv.Status, "actual") {
							if len(richText) > 0 {
								comma := &excelize.RichTextRun{
									Text: ",",
									Font: &excelize.Font{
										Bold:  true,
										Size:  10,
										Color: "000000",
									},
								}
								richText = append(richText, *comma)
							}
							tr := &excelize.RichTextRun{
								Text: lv.LeaveDate.Format("02 Jan"),
								Font: &excelize.Font{
									Bold:  true,
									Size:  10,
									Color: "000000",
								},
							}
							richText = append(richText, *tr)
							if lv.Hours < 8.0 {
								tr = &excelize.RichTextRun{
									Text: "(" + fmt.Sprintf("%.1f", lv.Hours) + ")",
									Font: &excelize.Font{
										Bold:      true,
										Size:      7,
										Color:     "000000",
										VertAlign: "superscript",
									},
								}
								richText = append(richText, *tr)
							}
						} else {
							if len(richText) > 0 {
								comma := &excelize.RichTextRun{
									Text: ",",
									Font: &excelize.Font{
										Bold:  true,
										Size:  10,
										Color: "000000",
									},
								}
								richText = append(richText, *comma)
							}
							tr := &excelize.RichTextRun{
								Text: lv.LeaveDate.Format("02 Jan"),
								Font: &excelize.Font{
									Bold:  true,
									Size:  10,
									Color: "3366ff",
								},
							}
							richText = append(richText, *tr)
							if lv.Hours < 8.0 {
								tr = &excelize.RichTextRun{
									Text: "(" + fmt.Sprintf("%.1f", lv.Hours) + ")",
									Font: &excelize.Font{
										Bold:      true,
										Size:      7,
										Color:     "3366ff",
										VertAlign: "superscript",
									},
								}
								richText = append(richText, *tr)
							}
						}
					}
				}
				lr.Report.SetCellRichText(sheetName, GetCellID(2, row+holRow), richText)
			}
			col = 4
		}
		sort.Sort(ByLeaveMonth(months))
		for _, month := range months {
			lvRow++
			style := lr.Styles["ptodates"]
			if month.Disable {
				style = lr.Styles["disabledlt"]
			}
			lr.Report.SetCellStyle(sheetName, GetCellID(col+0, row+lvRow),
				GetCellID(col+1, row+lvRow), style)
			lr.Report.MergeCell(sheetName, GetCellID(col+0, row+lvRow),
				GetCellID(col+1, row+lvRow))
			style = lr.Styles["ptotaken"]
			if month.Disable {
				style = lr.Styles["disabled"]
			}
			lr.Report.SetCellStyle(sheetName, GetCellID(col+2, row+lvRow),
				GetCellID(col+2, row+lvRow), style)
			lr.Report.SetCellValue(sheetName, GetCellID(col+2, row+lvRow),
				month.GetPTOActual())
			style = lr.Styles["ptorequest"]
			if month.Disable {
				style = lr.Styles["disabled"]
			}
			lr.Report.SetCellStyle(sheetName, GetCellID(col+3, row+lvRow),
				GetCellID(col+3, row+lvRow), style)
			lr.Report.SetCellValue(sheetName, GetCellID(col+3, row+lvRow),
				month.GetPTOSchedule())

			richText = richText[:0]
			bComma := false

			rt := &excelize.RichTextRun{
				Text: month.Month.Format("Jan") + ": ",
				Font: &excelize.Font{
					Bold:  true,
					Size:  10,
					Color: "ff0000",
				},
			}
			richText = append(richText, *rt)
			for _, prd := range month.Periods {
				if bComma {
					comma := &excelize.RichTextRun{
						Text: ",",
						Font: &excelize.Font{
							Bold:  true,
							Size:  10,
							Color: "000000",
						},
					}
					richText = append(richText, *comma)
				}
				wc := lr.Workcodes[prd.Code]
				text := ""
				if !prd.StartDate.Equal(prd.EndDate) {
					text = prd.StartDate.Format("2") + "-" +
						prd.EndDate.Format("2")
				} else {
					text = prd.StartDate.Format("2")
				}
				var color string
				if !strings.EqualFold(prd.Code, "v") ||
					!strings.EqualFold(prd.Status, "actual") {
					color = wc.BackColor
				} else {
					color = "000000"
				}
				rt = &excelize.RichTextRun{
					Text: text,
					Font: &excelize.Font{
						Bold:  true,
						Size:  10,
						Color: color,
					},
				}
				richText = append(richText, *rt)

				if len(prd.Leaves) == 1 && prd.Leaves[0].Hours < std {
					rt = &excelize.RichTextRun{
						Text: "(" + fmt.Sprintf("%.1f", prd.Leaves[0].Hours) + ")",
						Font: &excelize.Font{
							Bold:      true,
							Size:      10,
							Color:     color,
							VertAlign: "superscript",
						},
					}
					richText = append(richText, *rt)
				}
				bComma = true
			}
			lr.Report.SetCellRichText(sheetName, GetCellID(col, row+lvRow), richText)
		}

		if holRow > lvRow {
			row += holRow
		} else {
			row += lvRow
		}

		// Add totals labels and data rows
		row++
		col = 0
		if lr.BHolidays {
			style = lr.Styles["sectionlbl"]
			lr.Report.SetCellStyle(sheetName, GetCellID(0, row),
				GetCellID(3, row), style)
			lr.Report.SetCellValue(sheetName, GetCellID(0, row),
				"")
			lr.Report.SetCellValue(sheetName, GetCellID(1, row),
				"Days Left")
			lr.Report.SetCellValue(sheetName, GetCellID(2, row),
				"Hours Left")
			lr.Report.SetCellValue(sheetName, GetCellID(3, row),
				"Total Hours")
			daysLeft := 0
			hoursTaken := 0.0
			for _, hol := range lr.Holidays {
				if hol.GetHolidayHours() < 8.0 && !hol.Disable {
					daysLeft++
				}
				if hol.GetHolidayHours() > 0.0 {
					hoursTaken += hol.GetHolidayHours()
				}
			}
			hoursLeft := daysLeft * 8.0
			style = lr.Styles["holdaysleft"]
			lr.Report.SetCellStyle(sheetName, GetCellID(0, row+1),
				GetCellID(1, row+1), style)
			style = lr.Styles["hollblactual"]
			lr.Report.SetCellStyle(sheetName, GetCellID(2, row+1),
				GetCellID(3, row+1), style)
			lr.Report.SetCellValue(sheetName, GetCellID(0, row+1), "")
			lr.Report.SetCellValue(sheetName, GetCellID(1, row+1), daysLeft)
			lr.Report.SetCellValue(sheetName, GetCellID(2, row+1), hoursLeft)
			lr.Report.SetCellValue(sheetName, GetCellID(3, row+1), hoursTaken)
			col = 4
		}

		style = lr.Styles["sectionlbl"]
		lr.Report.SetCellStyle(sheetName, GetCellID(col+0, row),
			GetCellID(col+4, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col+0, row),
			"Annual Leave")
		lr.Report.SetCellValue(sheetName, GetCellID(col+1, row),
			"Carry")
		lr.Report.SetCellValue(sheetName, GetCellID(col+2, row),
			"Total Taken")
		lr.Report.SetCellValue(sheetName, GetCellID(col+3, row),
			"Request")
		lr.Report.SetCellValue(sheetName, GetCellID(col+4, row),
			"Balance")
		totalTaken := 0.0
		totalRequested := 0.0
		for _, mon := range months {
			totalTaken += mon.GetPTOActual()
			totalRequested += mon.GetPTOSchedule()
		}
		balance := (annual + carry) - (totalTaken + totalRequested)
		style = lr.Styles["hollblactual"]
		style2 := lr.Styles["ptorequest"]
		balstyle := lr.Styles["balance"]
		lr.Report.SetCellStyle(sheetName, GetCellID(col+0, row+1),
			GetCellID(col+2, row+1), style)
		lr.Report.SetCellStyle(sheetName, GetCellID(col+4, row+1),
			GetCellID(col+4, row+1), balstyle)
		lr.Report.SetCellStyle(sheetName, GetCellID(col+3, row+1),
			GetCellID(col+3, row+1), style2)
		lr.Report.SetCellValue(sheetName, GetCellID(col+0, row+1),
			annual)
		lr.Report.SetCellValue(sheetName, GetCellID(col+1, row+1),
			carry)
		lr.Report.SetCellValue(sheetName, GetCellID(col+2, row+1),
			totalTaken)
		lr.Report.SetCellValue(sheetName, GetCellID(col+3, row+1),
			totalRequested)
		lr.Report.SetCellValue(sheetName, GetCellID(col+4, row+1),
			balance)
		row += 2
	}

	return nil
}

func (lr *LeaveReport) CreateFullMonthlyReference() error {
	sheetName := strconv.Itoa(lr.Year) + " Monthly"
	lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)

	// set all the column widths
	lr.Report.SetColWidth(sheetName, "A", "A", 3.5)
	lr.Report.SetColWidth(sheetName, "B", "B", 20.0)
	lr.Report.SetColWidth(sheetName, "C", "AG", 4.5)
	lr.Report.SetColWidth(sheetName, "AH", "AJ", 7.0)

	// add page label at top, legend on left and freeze
	// the panes
	style := lr.Styles["currentasof"]
	lr.Report.SetCellStyle(sheetName, "A1", "D1", style)
	lr.Report.MergeCell(sheetName, "A1", "D1")
	lr.Report.SetCellValue(sheetName, "A1",
		"Current As of: "+time.Now().Format("01/02/2006"))

	style = lr.Styles["header"]
	lr.Report.SetCellStyle(sheetName, "E1", "AH1", style)
	lr.Report.MergeCell(sheetName, "E1", "AH1")
	lr.Report.SetCellValue(sheetName, "E1",
		strconv.Itoa(lr.Year)+" PTO/HOL Quick Reference (FULL)")

	row := 2
	col := 2
	workcodes := maps.Values(lr.Workcodes)
	sort.Sort(labor.ByWorkcode(workcodes))
	for _, v := range workcodes {
		style = lr.Styles[v.Id]
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col, row), "")
		style = lr.Styles["legend"]
		lr.Report.SetCellStyle(sheetName, GetCellID(col+1, row),
			GetCellID(col+5, row), style)
		lr.Report.MergeCell(sheetName, GetCellID(col+1, row),
			GetCellID(col+5, row))
		lr.Report.SetCellValue(sheetName, GetCellID(col+1, row),
			v.Title)
		col += 6
		if col > 31 {
			row++
			col = 2
		}
	}

	freezeCell := GetCellID(1, row+1)
	option := &excelize.Selection{
		SQRef:      freezeCell,
		ActiveCell: freezeCell,
		Pane:       "bottomLeft",
	}
	pane := &excelize.Panes{
		Freeze:      true,
		YSplit:      row,
		TopLeftCell: freezeCell,
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			*option,
		},
	}
	err := lr.Report.SetPanes(sheetName, pane)
	row++

	if err != nil {
		return err
	}

	// add months to report sheet
	current := time.Date(lr.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := current.AddDate(1, 0, 0)
	for current.Before(end) {
		row, err = lr.CreateQuickReferenceMonth(sheetName, current, row, true)
		if err != nil {
			return err
		}
		current = current.AddDate(0, 1, 0)
	}

	return nil
}

func (lr *LeaveReport) CreateMinumimMonthlyReference() error {
	sheetName := strconv.Itoa(lr.Year) + " Monthly (Minumum)"
	lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)

	// set all the column widths
	lr.Report.SetColWidth(sheetName, "A", "A", 3.5)
	lr.Report.SetColWidth(sheetName, "B", "B", 20.0)
	lr.Report.SetColWidth(sheetName, "C", "AG", 4.5)
	lr.Report.SetColWidth(sheetName, "AH", "AJ", 7.0)

	// add page label at top, legend on left and freeze
	// the panes
	style := lr.Styles["currentasof"]
	lr.Report.SetCellStyle(sheetName, "A1", "D1", style)
	lr.Report.MergeCell(sheetName, "A1", "D1")
	lr.Report.SetCellValue(sheetName, "A1",
		"Current As of: "+time.Now().Format("01/02/2006"))

	style = lr.Styles["header"]
	lr.Report.SetCellStyle(sheetName, "E1", "AH1", style)
	lr.Report.MergeCell(sheetName, "E1", "AH1")
	lr.Report.SetCellValue(sheetName, "E1",
		strconv.Itoa(lr.Year)+" PTO/HOL Quick Reference (FULL)")

	row := 2
	col := 2
	workcodes := maps.Values(lr.Workcodes)
	sort.Sort(labor.ByWorkcode(workcodes))
	for _, v := range workcodes {
		style = lr.Styles[v.Id]
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col, row), "")
		style = lr.Styles["legend"]
		lr.Report.SetCellStyle(sheetName, GetCellID(col+1, row),
			GetCellID(col+5, row), style)
		lr.Report.MergeCell(sheetName, GetCellID(col+1, row),
			GetCellID(col+5, row))
		lr.Report.SetCellValue(sheetName, GetCellID(col+1, row),
			v.Title)
		col += 6
		if col > 31 {
			row++
			col = 2
		}
	}

	freezeCell := GetCellID(1, row+1)
	option := &excelize.Selection{
		SQRef:      freezeCell,
		ActiveCell: freezeCell,
		Pane:       "bottomLeft",
	}
	pane := &excelize.Panes{
		Freeze:      true,
		YSplit:      row,
		TopLeftCell: freezeCell,
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			*option,
		},
	}
	err := lr.Report.SetPanes(sheetName, pane)
	row++

	if err != nil {
		return err
	}

	// add months to report sheet
	current := time.Date(lr.Year, 1, 1, 0, 0, 0, 0, time.UTC)
	end := current.AddDate(1, 0, 0)
	for current.Before(end) {
		row, err = lr.CreateQuickReferenceMonth(sheetName, current, row, false)
		if err != nil {
			return err
		}
		current = current.AddDate(0, 1, 0)
	}

	return nil
}

func (lr *LeaveReport) CreateQuickReferenceMonth(
	sheetname string, month time.Time, row int, full bool) (int, error) {
	row++

	// month, days of month, and other labels
	style := lr.Styles["monthup"]
	lr.Report.SetCellStyle(sheetname, GetCellID(1, row),
		GetCellID(1, row), style)
	lr.Report.SetCellValue(sheetname, GetCellID(1, row),
		month.Format("January"))
	style = lr.Styles["monthdn"]
	lr.Report.SetCellStyle(sheetname, GetCellID(1, row+1),
		GetCellID(1, row+1), style)
	lr.Report.SetCellValue(sheetname, GetCellID(1, row+1),
		month.Format("2006"))
	current := time.Date(month.Year(), month.Month(), 1, 0, 0,
		0, 0, time.UTC)
	end := current.AddDate(0, 1, 0)
	col := 2
	for current.Before(end) {
		sStyle := "weekday"
		if current.Weekday() == time.Saturday ||
			current.Weekday() == time.Sunday {
			sStyle = "weekend"
		}
		style = lr.Styles[sStyle+"up"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row),
			current.Format("_2"))
		style = lr.Styles[sStyle+"dn"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row+1),
			GetCellID(col, row+1), style)
		weekday := current.Format("Mon")
		lr.Report.SetCellValue(sheetname, GetCellID(col, row+1),
			weekday[:2])
		current = current.AddDate(0, 0, 1)
		col++
	}

	if full {
		col = 33
		style = lr.Styles["totalup"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row),
			"Total")
		style = lr.Styles["totaldn"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row+1),
			GetCellID(col, row+1), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row+1),
			"Hours")
		col++
		style = lr.Styles["holup"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row),
			"Hol/")
		style = lr.Styles["holdn"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row+1),
			GetCellID(col, row+1), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row+1),
			"Other")
		col++
		style = lr.Styles["ptoup"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row),
			"PTO")
		style = lr.Styles["ptodn"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row+1),
			GetCellID(col, row+1), style)
		lr.Report.SetCellValue(sheetname, GetCellID(col, row+1),
			"Only")
	}

	row++

	startRow := row
	for _, emp := range lr.Employees {
		if full || emp.GetLeaveHours(month, month.AddDate(0, 1, 0)) > 0.0 {
			row++
			lr.CreateEmployeeRow(sheetname, month, emp, row, full)
		}
	}
	endRow := row

	if full {
		row++
		col = 33
		formula := "SUM(" + GetCellID(col, startRow) + ":" +
			GetCellID(col, endRow) + ")"
		style = lr.Styles["balance"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellFormula(sheetname, GetCellID(col, row),
			formula)
		col++
		style = lr.Styles["weekday"]
		lr.Report.SetCellStyle(sheetname, GetCellID(col, row),
			GetCellID(col+1, row), style)
		lr.Report.MergeCell(sheetname, GetCellID(col, row),
			GetCellID(col+1, row))
		lr.Report.SetCellValue(sheetname, GetCellID(col, row),
			"TOTALS")
	}
	row++

	return row, nil
}

func (lr *LeaveReport) CreateEmployeeRow(sheetName string,
	month time.Time, emp employees.Employee, row int,
	full bool) {
	col := 1
	current := time.Date(month.Year(), month.Month(), 1, 0, 0,
		0, 0, time.UTC)
	end := current.AddDate(0, 1, 0)
	style := lr.Styles["weekday"]
	lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
		GetCellID(col, row), style)
	lr.Report.SetCellValue(sheetName, GetCellID(col, row),
		emp.Name.GetLastFirst())
	col++
	labor := make([]employees.EmployeeLaborCode, 0)
	for current.Before(end) {
		wd := emp.GetWorkdayActual(current, labor)
		sStyle := ""
		display := 0.0
		for _, wc := range lr.Workcodes {
			if wd != nil {
				if strings.EqualFold(wc.Id, wd.Code) && wc.IsLeave {
					sStyle = wc.Id
					display = wd.Hours
				}
			}
		}
		if sStyle == "" {
			sStyle = "weekday"
			if current.Weekday() == time.Saturday ||
				current.Weekday() == time.Sunday {
				sStyle = "weekend"
			}
		}
		style = lr.Styles[sStyle]
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		if display > 0.0 {
			lr.Report.SetCellValue(sheetName, GetCellID(col, row),
				display)
		} else {
			lr.Report.SetCellValue(sheetName, GetCellID(col, row),
				"")
		}
		current = current.AddDate(0, 0, 1)
		col++
	}
	if full {
		col = 33
		totalHrs := emp.GetLeaveHours(month, end)
		ptoHrs := emp.GetPTOHours(month, end)
		otherHrs := totalHrs - ptoHrs
		style = lr.Styles["weekday"]
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col, row),
			totalHrs)
		col++
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col, row),
			otherHrs)
		col++
		lr.Report.SetCellStyle(sheetName, GetCellID(col, row),
			GetCellID(col, row), style)
		lr.Report.SetCellValue(sheetName, GetCellID(col, row),
			ptoHrs)
	}
}
