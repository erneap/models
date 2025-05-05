package reports

import (
	"math"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/go-models/employees"
	"github.com/erneap/go-models/labor"
	"github.com/erneap/go-models/sites"
	"github.com/erneap/go-models/svcs"
	"github.com/xuri/excelize/v2"
	"golang.org/x/exp/maps"
)

type LaborReport struct {
	Report            *excelize.File
	Date              time.Time
	TeamID            string
	SiteID            string
	CompanyID         string
	ForecastReports   []sites.ForecastReport
	Workcodes         map[string]labor.Workcode
	Styles            map[string]int
	ConditionalStyles map[string]int
	Employees         []employees.Employee
	StatsRow          int
	CurrentAsOf       time.Time
	EndWork           time.Time
	Offset            float64
}

func (lr *LaborReport) Create() error {
	lr.CurrentAsOf = time.Now().UTC()
	lr.EndWork = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	lr.StatsRow = 3
	lr.Styles = make(map[string]int)
	lr.ConditionalStyles = make(map[string]int)
	lr.Workcodes = make(map[string]labor.Workcode)
	lr.Report = excelize.NewFile()

	// Get list of forecast reports for the team/site
	minDate := time.Date(lr.Date.Year(), lr.Date.Month(),
		lr.Date.Day(), 0, 0, 0, 0, time.UTC)
	maxDate := time.Date(lr.Date.Year(), lr.Date.Month(),
		lr.Date.Day(), 0, 0, 0, 0, time.UTC)
	site, err := svcs.GetSite(lr.TeamID, lr.SiteID)
	if err != nil {
		return err
	}
	for _, fr := range site.ForecastReports {
		if strings.EqualFold(lr.CompanyID, fr.CompanyID) &&
			(lr.Date.Equal(fr.StartDate) ||
				lr.Date.Equal(fr.EndDate) ||
				(lr.Date.After(fr.StartDate) &&
					lr.Date.Before(fr.EndDate))) {
			sort.Sort(labor.ByLaborCode(fr.LaborCodes))
			lr.ForecastReports = append(lr.ForecastReports, fr)
			if fr.StartDate.Before(minDate) {
				minDate = time.Date(fr.StartDate.Year(),
					fr.StartDate.Month(), fr.StartDate.Day(), 0, 0, 0,
					0, time.UTC)
			}
			if fr.EndDate.After(maxDate) {
				maxDate = time.Date(fr.EndDate.Year(),
					fr.EndDate.Month(), fr.EndDate.Day(), 0, 0, 0,
					0, time.UTC)
			}
		}
	}

	// Get the team's workcodes
	team, err := svcs.GetTeam(lr.TeamID)
	if err != nil {
		return err
	}
	for _, wc := range team.Workcodes {
		lr.Workcodes[wc.Id] = wc
	}
	for _, site := range team.Sites {
		if strings.EqualFold(site.ID, lr.SiteID) {
			lr.Offset = 0
		}
	}

	// get employees with assignments for the site that are assigned
	// during the forecast period.

	emps, err := svcs.GetEmployeesForTeam(lr.TeamID)
	if err != nil {
		return err
	}
	for _, emp := range emps {
		if emp.AtSite(lr.SiteID, minDate, maxDate) {
			// get work records for the year inclusive of the two
			// dates
			work, _ := svcs.GetEmployeeWork(emp.ID.Hex(),
				uint(minDate.Year()))
			if work != nil {
				emp.Work = append(emp.Work, work.Work...)
			}
			if minDate.Year() != maxDate.Year() {
				year := minDate.Year()
				for year < maxDate.Year() {
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
			lr.Employees = append(lr.Employees, emp)
		}
		lr.Report.UpdateLinkedValue()
	}

	//////////////////////////////////////////////////////////
	// Report Creation
	//////////////////////////////////////////////////////////
	lr.CreateStyles()

	lr.CreateStatisticsReport()

	sort.Sort(sites.ByForecastReport(lr.ForecastReports))

	for _, fr := range lr.ForecastReports {
		// current report
		lr.CreateContractReport(fr, true)

		// forecast report
		lr.CreateContractReport(fr, false)
	}

	lr.Report.DeleteSheet("Sheet1")

	return nil
}

func (lr *LaborReport) CreateStyles() error {
	numFmt := "##0.0;[Red]##0.0;-;@"
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

func (lr *LaborReport) CreateContractReport(
	fr sites.ForecastReport, current bool) error {
	sheetName := fr.Name + "_"
	if current {
		sheetName += "Current"
	} else {
		sheetName += "Forecast"
	}
	lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)

	// set column widths
	lr.Report.SetColWidth(sheetName, "A", "A", 8.0)
	lr.Report.SetColWidth(sheetName, "B", "B", 10.57)
	lr.Report.SetColWidth(sheetName, "C", "C", 9.0)
	lr.Report.SetColWidth(sheetName, "D", "D", 14.57)
	lr.Report.SetColWidth(sheetName, "E", "E", 10.57)
	lr.Report.SetColWidth(sheetName, "F", "F", 20.0)
	lr.Report.SetColWidth(sheetName, "G", "G", 18.14)
	lr.Report.SetColWidth(sheetName, "H", "H", 17.57)
	lr.Report.SetColWidth(sheetName, "I", "K", 14.57)
	lr.Report.SetColWidth(sheetName, "L", "L", 55.14)
	columns := 12
	for _, period := range fr.Periods {
		sumCol := GetColumn(columns)
		startCol := GetColumn(columns + 1)
		endCol := GetColumn(columns + len(period.Periods))
		lr.Report.SetColWidth(sheetName, sumCol, sumCol, 12.00)
		lr.Report.SetColWidth(sheetName, startCol, endCol, 9.00)
		columns += len(period.Periods) + 1
	}
	lr.Report.SetColWidth(sheetName, GetColumn(columns), GetColumn(columns), 15.43)

	// headers for page
	style := lr.Styles["header"]
	lr.Report.SetCellStyle(sheetName, "A1", "G1", style)
	lr.Report.MergeCell(sheetName, "A1", "G1")
	lr.Report.SetCellValue(sheetName, "A1", "FFP Labor: "+
		"CLIN "+fr.LaborCodes[0].CLIN+" SUMMARY")

	lr.Report.SetCellStyle(sheetName, "A2", "G2", style)
	lr.Report.MergeCell(sheetName, "A2", "G2")
	lr.Report.SetCellValue(sheetName, "A2", fr.Name+
		"Year - "+fr.StartDate.Format("02 Jan 06")+" "+
		fr.EndDate.Format("02 Jan 06"))

	style = lr.Styles["label"]
	lr.Report.SetCellStyle(sheetName, "L1", "L1", style)
	lr.Report.SetCellValue(sheetName, "L1",
		"Weeks Per Accounting Month")
	lr.Report.SetCellStyle(sheetName, "L2", "L2", style)
	lr.Report.SetCellValue(sheetName, "L2",
		"Accounting Month")
	lr.Report.SetCellStyle(sheetName, "L3", "L3", style)
	lr.Report.SetCellValue(sheetName, "L3",
		"Week Ending")

	style = lr.Styles["peopleheader"]
	lr.Report.SetCellStyle(sheetName, "A4", "L4", style)
	lr.Report.SetCellValue(sheetName, "A4", "CLIN")
	lr.Report.SetCellValue(sheetName, "B4", "SLIN")
	lr.Report.SetCellValue(sheetName, "C4", "Company")
	lr.Report.SetCellValue(sheetName, "D4", "Location")
	lr.Report.SetCellValue(sheetName, "E4", "WBS")
	lr.Report.SetCellValue(sheetName, "F4", "Labor NWA")
	lr.Report.SetCellValue(sheetName, "G4", "Last Name")
	lr.Report.SetCellValue(sheetName, "H4", "Labor Category")
	lr.Report.SetCellValue(sheetName, "I4", "Employee ID")
	lr.Report.SetCellValue(sheetName, "J4", "PeopleSoft ID")
	lr.Report.SetCellValue(sheetName, "K4", "Cost Center")
	lr.Report.SetCellValue(sheetName, "L4", "Comments/Remarks")

	// set period for the report sheet
	column := 12
	for _, period := range fr.Periods {
		// For each monthly period it will consist of a month summary and weekly periods
		style = lr.Styles["periods"]
		cellID := GetCellID(column, 1)
		lr.Report.SetCellStyle(sheetName, cellID, cellID,
			style)
		lr.Report.SetCellValue(sheetName, cellID,
			strconv.Itoa(len(period.Periods)))
		style = lr.Styles["month"]
		lr.Report.SetCellStyle(sheetName, GetCellID(column, 2),
			GetCellID(column+len(period.Periods), 2), style)
		lr.Report.MergeCell(sheetName, GetCellID(column, 2),
			GetCellID(column+len(period.Periods), 2))
		lr.Report.SetCellValue(sheetName, GetCellID(column, 2),
			strings.ToUpper(period.Month.Format("January")))
		style = lr.Styles["monthsumlbl"]
		lr.Report.SetCellStyle(sheetName, GetCellID(column, 3),
			GetCellID(column, 3), style)
		lr.Report.SetCellValue(sheetName, GetCellID(column, 3),
			"Month Total")
		style = lr.Styles["monthdate"]
		cellID = GetCellID(column, 4)
		lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
		lr.Report.SetCellValue(sheetName, cellID,
			period.Month.Format("Jan-06"))
		lr.Report.SetColOutlineLevel(sheetName, GetColumn(column), 0)
		style = lr.Styles["dates"]
		style2 := lr.Styles["weeks"]
		for i, prd := range period.Periods {
			cellID := GetCellID(column+i+1, 3)
			lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
			lr.Report.SetCellValue(sheetName, cellID, prd.Format("02-Jan"))
			cellID = GetCellID(column+i+1, 4)
			lr.Report.SetCellStyle(sheetName, cellID, cellID, style2)
			lr.Report.SetCellValue(sheetName, cellID, "Week"+strconv.Itoa(i+1))
			lr.Report.SetColOutlineLevel(sheetName, GetColumn(column+i+1), 1)
		}
		column += len(period.Periods) + 1
	}
	// add column for totals
	style = lr.Styles["monthsumlbl"]
	cellID := GetCellID(column, 4)
	lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
	lr.Report.SetCellValue(sheetName, cellID, "EAC")
	lr.Report.SetColOutlineLevel(sheetName, GetColumn(column), 0)
	lastWorkday := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	sort.Sort(employees.ByEmployees(lr.Employees))
	for _, emp := range lr.Employees {
		last := emp.GetLastWorkday()
		if last.After(lastWorkday) {
			lastWorkday = last
		}
	}

	if fr.SortByFirst {
		sort.Sort(employees.ByEmployeesFirst(lr.Employees))
	} else {
		sort.Sort(employees.ByEmployees(lr.Employees))
	}

	// create a list of leave/work codes for comparison help determine working or
	// not working criteria
	row := 4
	var compareCodes []employees.EmployeeCompareCode
	for _, wc := range maps.Values(lr.Workcodes) {
		cc := &employees.EmployeeCompareCode{
			Code:    wc.Id,
			IsLeave: wc.IsLeave,
		}
		compareCodes = append(compareCodes, *cc)
	}

	// step through labor codes for report
	for _, lCode := range fr.LaborCodes {

		// step thorugh the employees to see if they worked or are forecast to use
		// the labor code
		for _, emp := range lr.Employees {
			actual := emp.GetWorkedHoursForLabor(
				lCode.ChargeNumber, lCode.Extension, fr.StartDate,
				fr.EndDate.AddDate(0, 0, 1))
			forecast := emp.GetForecastHours(lCode, fr.StartDate,
				fr.EndDate.AddDate(0, 0, 1),
				compareCodes, lr.Offset)
			if actual > 0.0 || forecast > 0.0 {
				// show employee for this labor code, if either actual or forecast > 0
				row++
				style = lr.Styles["peoplectr"]
				lStyle := lr.Styles["peopleleft"]
				if strings.Contains(emp.CompanyInfo.JobTitle, "Liaison") {
					style = lr.Styles["liaisonctr"]
					lStyle = lr.Styles["liaisonleft"]
				}
				lr.Report.SetCellStyle(sheetName, GetCellID(0, row), GetCellID(5, row),
					style)
				lr.Report.SetCellStyle(sheetName, GetCellID(6, row), GetCellID(6, row),
					lStyle)
				lr.Report.SetCellStyle(sheetName, GetCellID(7, row), GetCellID(11, row),
					style)
				lr.Report.SetCellValue(sheetName, GetCellID(0, row), lCode.CLIN)
				lr.Report.SetCellValue(sheetName, GetCellID(1, row), lCode.SLIN)
				lr.Report.SetCellValue(sheetName, GetCellID(2, row),
					strings.ToUpper(emp.CompanyInfo.Division))
				lr.Report.SetCellValue(sheetName, GetCellID(3, row), lCode.Location)
				lr.Report.SetCellValue(sheetName, GetCellID(4, row), lCode.WBS)
				lr.Report.SetCellValue(sheetName, GetCellID(5, row),
					lCode.ChargeNumber+" "+lCode.Extension)
				lr.Report.SetCellValue(sheetName, GetCellID(6, row), emp.Name.LastName)
				lr.Report.SetCellValue(sheetName, GetCellID(7, row),
					emp.CompanyInfo.Rank)
				lr.Report.SetCellValue(sheetName, GetCellID(8, row),
					emp.CompanyInfo.EmployeeID)
				lr.Report.SetCellValue(sheetName, GetCellID(9, row),
					emp.CompanyInfo.AlternateID)
				lr.Report.SetCellValue(sheetName, GetCellID(10, row),
					emp.CompanyInfo.CostCenter)
				column = 11
				var sumlist = []string{}

				// create columns for employee for this labor code (either worked or forecast)
				for _, period := range fr.Periods {
					column++
					style = lr.Styles["sum"]
					formula := ""
					if len(period.Periods) > 1 {
						formula = "=SUM(" + GetCellID(column+1, row) + ":" +
							GetCellID(column+(len(period.Periods)), row) + ")"
					} else {
						formula = "=" + GetCellID(column+1, row)
					}
					cellID = GetCellID(column, row)
					sumlist = append(sumlist, cellID)
					lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
					lr.Report.SetCellFormula(sheetName, cellID, formula)
					hours := 0.0
					for _, prd := range period.Periods {
						column++
						cellID = GetCellID(column, row)
						last := time.Date(prd.Year(), prd.Month(), prd.Day()+1, 0, 0, 0,
							0, time.UTC)
						first := last.AddDate(0, 0, -7)
						if first.Before(fr.StartDate) {
							first = time.Date(fr.StartDate.Year(), fr.StartDate.Month(),
								fr.StartDate.Day(), 0, 0, 0, 0, time.UTC)
						}
						if last.After(fr.EndDate.AddDate(0, 0, 1)) {
							last = time.Date(fr.EndDate.Year(), fr.EndDate.Month(),
								fr.EndDate.Day()+1, 0, 0, 0, 0, time.UTC)
						}
						style = lr.Styles["actual"]
						hours = emp.GetWorkedHoursForLabor(lCode.ChargeNumber,
							lCode.Extension, first, last)
						if !current {
							if last.AddDate(0, 0, -1).After(lastWorkday) {
								style = lr.Styles["forecast"]
							}
							hours += emp.GetForecastHours(lCode, first, last, compareCodes, lr.Offset)
						}
						lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
						format := lr.ConditionalStyles["cellpink"]
						lr.Report.SetConditionalFormat(sheetName, cellID,
							[]excelize.ConditionalFormatOptions{
								{Type: "cell", Criteria: "==", Format: format, Value: "0"},
							})
						lr.Report.SetCellValue(sheetName, cellID, hours)
					}
				}
				style = lr.Styles["monthsum"]
				column++
				cellID = GetCellID(column, row)
				formula := ""
				for _, val := range sumlist {
					if formula == "" {
						formula += "="
					} else {
						formula += "+"
					}
					formula += val
				}
				lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
				format := lr.ConditionalStyles["pinkright"]
				lr.Report.SetConditionalFormat(sheetName, cellID,
					[]excelize.ConditionalFormatOptions{
						{Type: "cell", Criteria: ">",
							Format: format,
							Value:  strconv.FormatFloat(lCode.HoursPerEmployee, 'f', 1, 64),
						},
					})
				format = lr.ConditionalStyles["greenedright"]
				lr.Report.SetConditionalFormat(sheetName, cellID,
					[]excelize.ConditionalFormatOptions{
						{Type: "cell", Criteria: "<=",
							Format: format,
							Value:  strconv.FormatFloat(lCode.HoursPerEmployee, 'f', 1, 64),
						},
					})
				lr.Report.SetCellFormula(sheetName, cellID, formula)
			}
		}

		// create statistics page data but only for actual data sheets
		if current {
			days := math.Ceil(lCode.EndDate.Sub(lCode.StartDate).Hours() / 24.0)
			daysToNow := math.Ceil(lr.EndWork.Sub(lCode.StartDate).Hours() / 24.0)
			lr.StatsRow += 1
			codeTxt := lCode.ChargeNumber + " " + lCode.Extension
			totalHours := lCode.HoursPerEmployee * float64(lCode.MinimumEmployees)
			perDay := totalHours / days
			codeHours := perDay * daysToNow
			if codeHours > totalHours {
				codeHours = totalHours
			}
			if codeHours < 0.0 {
				codeHours = 0.0
			}
			pctStyle := 0

			bLight := (lr.StatsRow%2 == 0)
			if bLight {
				style = lr.Styles["graytext"]
			} else {
				style = lr.Styles["whitetext"]
			}
			lr.Report.SetCellStyle("Statistics", GetCellID(0,
				lr.StatsRow), GetCellID(2, lr.StatsRow), style)
			lr.Report.SetCellValue("Statistics", GetCellID(0,
				lr.StatsRow), codeTxt)
			lr.Report.SetCellValue("Statistics", GetCellID(1,
				lr.StatsRow), lCode.StartDate.Format("01/02/2006"))
			lr.Report.SetCellValue("Statistics", GetCellID(2,
				lr.StatsRow), lCode.EndDate.Format("01/02/2006"))
			if bLight {
				style = lr.Styles["bluenum"]
				pctStyle = lr.Styles["bluepct"]
			} else {
				style = lr.Styles["ltbluenum"]
				pctStyle = lr.Styles["ltbluepct"]
			}
			lr.Report.SetCellStyle("Statistics", GetCellID(3,
				lr.StatsRow), GetCellID(5, lr.StatsRow), style)
			lr.Report.SetCellStyle("Statistics", GetCellID(6,
				lr.StatsRow), GetCellID(6, lr.StatsRow), pctStyle)
			lr.Report.SetCellValue("Statistics", GetCellID(3,
				lr.StatsRow), codeHours)
			formula := "=SUMIF(" + sheetName + "!" +
				GetCellID(5, 5) + ":" + GetCellID(5, row) + ", \"*" +
				lCode.ChargeNumber + " " + lCode.Extension + "*\", " +
				sheetName + "!" + GetCellID(column, 5) + ":" +
				GetCellID(column, row) + ")"
			lr.Report.SetCellFormula("Statistics", GetCellID(4,
				lr.StatsRow), formula)
			formula = "=" + GetCellID(4, lr.StatsRow) + "-" +
				GetCellID(3, lr.StatsRow)
			lr.Report.SetCellFormula("Statistics", GetCellID(5,
				lr.StatsRow), formula)
			formula = "=IFERROR(" + GetCellID(5, lr.StatsRow) +
				"/" + GetCellID(3, lr.StatsRow) + ",\"N/A\")"
			lr.Report.SetCellFormula("Statistics",
				GetCellID(6, lr.StatsRow), formula)
			if bLight {
				style = lr.Styles["greennum"]
				pctStyle = lr.Styles["greenpct"]
			} else {
				style = lr.Styles["ltgreennum"]
				pctStyle = lr.Styles["ltgreenpct"]
			}

			lr.Report.SetCellStyle("Statistics", GetCellID(7,
				lr.StatsRow), GetCellID(9, lr.StatsRow), style)
			lr.Report.SetCellStyle("Statistics", GetCellID(10,
				lr.StatsRow), GetCellID(10, lr.StatsRow), pctStyle)
			lr.Report.SetCellValue("Statistics", GetCellID(7,
				lr.StatsRow), totalHours)
			formula = "=SUMIF(" + fr.Name + "_Forecast!" +
				GetCellID(5, 5) + ":" + GetCellID(5, row) + ", \"*" +
				lCode.ChargeNumber + " " + lCode.Extension + "*\", " +
				fr.Name + "_Forecast!" + GetCellID(column, 5) + ":" +
				GetCellID(column, row) + ")"
			lr.Report.SetCellFormula("Statistics", GetCellID(8,
				lr.StatsRow), formula)
			formula = "=" + GetCellID(8, lr.StatsRow) + "-" +
				GetCellID(7, lr.StatsRow)
			lr.Report.SetCellFormula("Statistics", GetCellID(9,
				lr.StatsRow), formula)
			formula = "=IFERROR(" + GetCellID(9, lr.StatsRow) +
				"/" + GetCellID(7, lr.StatsRow) + ",\"N/A\")"
			lr.Report.SetCellFormula("Statistics",
				GetCellID(10, lr.StatsRow), formula)
		}
	}

	// add totals for each month summary and period summary
	style = lr.Styles["monthsumlbl"]
	cellID = GetCellID(3, row+1)
	lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
	lr.Report.SetCellValue(sheetName, cellID, "TOTAL")
	column = 11
	for _, period := range fr.Periods {
		column++
		formula := "=SUM(" + GetCellID(column, 5) + ":" +
			GetCellID(column, row) + ")"
		style = lr.Styles["monthsum"]
		cellID = GetCellID(column, row+1)
		lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
		lr.Report.SetCellFormula(sheetName, cellID, formula)
		for i := 0; i < len(period.Periods); i++ {
			column++
			formula = "=SUM(" + GetCellID(column, 5) + ":" +
				GetCellID(column, row) + ")"
			style = lr.Styles["cellsum"]
			cellID = GetCellID(column, row+1)
			lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
			lr.Report.SetCellFormula(sheetName, cellID, formula)
			format := lr.ConditionalStyles["cellpink"]
			lr.Report.SetConditionalFormat(sheetName, cellID,
				[]excelize.ConditionalFormatOptions{
					{Type: "cell", Criteria: "==", Format: format, Value: "0"},
				})
		}
	}
	column++
	style = lr.Styles["monthsum"]
	cellID = GetCellID(column, row+1)
	formula := "=SUM(" + GetCellID(column, 5) + ":" +
		GetCellID(column, row) + ")"
	lr.Report.SetCellStyle(sheetName, cellID, cellID, style)
	lr.Report.SetCellFormula(sheetName, cellID, formula)

	// hide columns A-E and H-L
	lr.Report.SetColVisible(sheetName, "A:E", false)
	lr.Report.SetColVisible(sheetName, "H:L", false)

	// freeze pane at column G (Name)
	lr.Report.SetPanes(sheetName, &excelize.Panes{
		Freeze:      true,
		XSplit:      7,
		YSplit:      4,
		TopLeftCell: "H5",
		ActivePane:  "bottomLeft",
		Selection: []excelize.Selection{
			{SQRef: "H5", ActiveCell: "H5", Pane: "bottomLeft"},
		},
	})
	return nil
}

func (lr *LaborReport) CreateStatisticsReport() error {
	sheetName := "Statistics"
	ind, _ := lr.Report.NewSheet(sheetName)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	lr.Report.SetSheetView(sheetName, 0, &options)
	lr.Report.SetActiveSheet(ind)

	// set column widths
	lr.Report.SetColWidth(sheetName, GetColumn(0),
		GetColumn(0), 30.0)
	lr.Report.SetColWidth(sheetName, GetColumn(1),
		GetColumn(2), 12.0)
	lr.Report.SetColWidth(sheetName, GetColumn(3),
		GetColumn(10), 12.0)

	// Current as of header
	style := lr.Styles["headerctr"]
	lr.Report.MergeCell(sheetName, "A1", "K1")
	lr.Report.SetCellStyle(sheetName, "A1", "A1", style)
	lr.Report.SetCellValue(sheetName, "A1", "Current As Of "+
		lr.CurrentAsOf.Format("02 Jan 2006"))

	lr.Report.SetRowHeight(sheetName, 1, 20.0)
	lr.Report.SetRowHeight(sheetName, 2, 77.0)
	style = lr.Styles["slantwhitelbl"]
	lr.Report.SetCellStyle(sheetName, "B2", "C2", style)
	lr.Report.SetCellValue(sheetName, "B2", "Start")
	lr.Report.SetCellValue(sheetName, "C2", "End")
	style = lr.Styles["slantbluelbl"]
	lr.Report.SetCellStyle(sheetName, "D2", "G2", style)
	lr.Report.SetCellValue(sheetName, "D2", "Alloted")
	lr.Report.SetCellValue(sheetName, "E2", "Used")
	lr.Report.SetCellValue(sheetName, "F2", "Over/Under")
	lr.Report.SetCellValue(sheetName, "G2", "Percent")
	style = lr.Styles["slantgreenlbl"]
	lr.Report.SetCellStyle(sheetName, "H2", "K2", style)
	lr.Report.SetCellValue(sheetName, "H2", "Alloted")
	lr.Report.SetCellValue(sheetName, "I2", "Used/Projected")
	lr.Report.SetCellValue(sheetName, "J2", "Over/Under")
	lr.Report.SetCellValue(sheetName, "K2", "Percent")
	style = lr.Styles["whitelbl"]
	lr.Report.SetCellStyle(sheetName, "A2", "A2", style)
	lr.Report.SetCellStyle(sheetName, "A3", "C3", style)
	lr.Report.SetCellValue(sheetName, "A3", "Contract No/Ext")
	lr.Report.MergeCell(sheetName, "B3", "C3")
	lr.Report.SetCellValue(sheetName, "B3", "Contract Period")
	style = lr.Styles["bluelbl"]
	lr.Report.SetCellStyle(sheetName, "D3", "G3", style)
	lr.Report.MergeCell(sheetName, "D3", "G3")
	lr.Report.SetCellValue(sheetName, "D3", "Current")
	style = lr.Styles["greenlbl"]
	lr.Report.SetCellStyle(sheetName, "H3", "K3", style)
	lr.Report.MergeCell(sheetName, "H3", "K3")
	lr.Report.SetCellValue(sheetName, "H3", "Forecast")

	return nil
}
