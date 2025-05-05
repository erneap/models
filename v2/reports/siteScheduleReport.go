package reports

import (
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/models/employees"
	"github.com/erneap/models/labor"
	"github.com/erneap/models/sites"
	"github.com/erneap/models/svcs"
	"github.com/xuri/excelize/v2"
)

type SiteScheduleReport struct {
	Report      *excelize.File
	Date        time.Time
	TeamID      string
	SiteID      string
	Workcenters []sites.Workcenter
	Workcodes   map[string]bool
	Styles      map[string]int
	Employees   []employees.Employee
}

func (sr *SiteScheduleReport) Create() error {
	sr.Styles = make(map[string]int)
	sr.Workcodes = make(map[string]bool)
	sr.Report = excelize.NewFile()
	sr.Date = time.Now().UTC()

	// get employees with assignments for the site that are assigned
	// during the year.
	startDate := time.Date(sr.Date.Year(), sr.Date.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 2, 0).AddDate(0, 0, -1)
	emps, err := svcs.GetEmployeesForTeam(sr.TeamID)
	if err != nil {
		return err
	}

	for _, emp := range emps {
		if emp.AtSite(sr.SiteID, startDate, endDate) {
			// get timecard data/work hours for each employee
			// for time period.
			wr, err := svcs.GetEmployeeWork(emp.ID.Hex(), uint(startDate.Year()))
			if err == nil {
				emp.Work = append(emp.Work, wr.Work...)
			}
			if startDate.Year() != endDate.Year() {
				wr, err = svcs.GetEmployeeWork(emp.ID.Hex(), uint(endDate.Year()))
				if err == nil {
					emp.Work = append(emp.Work, wr.Work...)
				}
			}
			sr.Employees = append(sr.Employees, emp)
		}
	}

	// get the site's workcenters
	site, err := svcs.GetSite(sr.TeamID, sr.SiteID)
	if err != nil {
		return err
	}
	sr.Workcenters = append(sr.Workcenters, site.Workcenters...)
	sort.Sort(sites.ByWorkcenter(sr.Workcenters))

	// create styles for display on each monthly sheet
	err = sr.CreateStyles()
	if err != nil {
		return err
	}

	// create monthly schedule for each month of the year
	for startDate.Before(endDate) {
		err = sr.AddMonth(startDate)
		if err != nil {
			return err
		}
		startDate = startDate.AddDate(0, 1, 0)
	}

	// add a leave legend sheet
	sr.CreateLegendSheet()

	// remove the provided sheet "Sheet1" from the workbook
	sr.Report.DeleteSheet("Sheet1")

	// save the file
	filename := "/tmp/siteschedule.xlsx"
	err = sr.Report.SaveAs(filename)

	// reload the file
	if err == nil {
		sr.Report, _ = excelize.OpenFile(filename)
	}

	return nil
}

func (sr *SiteScheduleReport) CreateStyles() error {
	//get all the workcodes from the team object and create the
	// styles for each one, plus one for weekend (non-leave), and
	// even and odd non-leaves.  Also need style for month label and workcenter

	team, err := svcs.GetTeam(sr.TeamID)
	if err != nil {
		return err
	}

	for _, wc := range team.Workcodes {
		style, err := sr.Report.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{Type: "pattern", Color: []string{wc.BackColor}, Pattern: 1},
			Font: &excelize.Font{Bold: true, Size: 11, Color: wc.TextColor},
			Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
				WrapText: true},
		})
		if err != nil {
			return err
		}
		sr.Styles[wc.Id] = style
		sr.Workcodes[wc.Id] = strings.EqualFold(wc.BackColor, "FFFFFF")
	}

	style, err := sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"C0C0C0"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["evenday"] = style

	style, err = sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"CCFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["weekend"] = style

	style, err = sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"00E6E6"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["evenend"] = style

	style, err = sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["weekday"] = style

	style, err = sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DE5D12"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["month"] = style

	style, err = sr.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 11, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	sr.Styles["workcenter"] = style
	return nil
}

func (sr *SiteScheduleReport) AddMonth(month time.Time) error {
	for w, wc := range sr.Workcenters {
		for s, sft := range wc.Shifts {
			sft.Employees = sft.Employees[:0]
			wc.Shifts[s] = sft
		}
		for p, pos := range wc.Positions {
			pos.Employees = pos.Employees[:0]
			wc.Positions[p] = pos
		}
		sr.Workcenters[w] = wc
	}
	startDate := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	for _, emp := range sr.Employees {
		if emp.AtSite(sr.SiteID, startDate, endDate) {
			// determine if employee assigned to position
			position := false
			for w, wc := range sr.Workcenters {
				for p, pos := range wc.Positions {
					for _, asgn := range pos.Assigned {
						if emp.ID.Hex() == asgn {
							position = true
							pos.Employees = append(pos.Employees, emp)
						}
					}
					if position {
						wc.Positions[p] = pos
						sr.Workcenters[w] = wc
					}
				}
			}
			if !position {
				wkctr, shift := emp.GetAssignment(startDate, endDate)
				for w, wc := range sr.Workcenters {
					if strings.EqualFold(wc.ID, wkctr) {
						for s, sft := range wc.Shifts {
							bShift := false
							for _, code := range sft.AssociatedCodes {
								if strings.EqualFold(code, shift) {
									bShift = true
								}
							}
							if bShift {
								sft.Employees = append(sft.Employees, emp)
								wc.Shifts[s] = sft
								sr.Workcenters[w] = wc
							}
						}
					}
				}
			}
		}
	}

	// create sheet for the month
	sheetLabel := startDate.Format("Jan06")
	var (
		size          = 1
		orientation   = "landscape"
		fitToHeight   = 1
		fitToWidth    = 1
		blackAndWhite = false
		fitToPage     = true
	)

	sr.Report.NewSheet(sheetLabel)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	sr.Report.SetSheetView(sheetLabel, 0, &options)
	if err := sr.Report.SetPageLayout(sheetLabel,
		&excelize.PageLayoutOptions{
			Size:          &size,
			Orientation:   &orientation,
			FitToHeight:   &fitToHeight,
			FitToWidth:    &fitToWidth,
			BlackAndWhite: &blackAndWhite,
		}); err != nil {
		return err
	}
	sr.Report.SetSheetProps(sheetLabel, &excelize.SheetPropsOptions{
		FitToPage: &fitToPage,
	})

	// set all the column widths
	sr.Report.SetColWidth(sheetLabel, "A", "A", 17.0)
	endOfMonth := endDate.AddDate(0, 0, -1)
	endColumn := "AE"
	switch endOfMonth.Day() {
	case 28:
		endColumn = "AC"
	case 29:
		endColumn = "AD"
	case 31:
		endColumn = "AF"
	default:
		endColumn = "AE"
	}
	sr.Report.SetColWidth(sheetLabel, "B", endColumn, 4.0)

	// monthly headers to include month label and days of the month
	style := sr.Styles["month"]
	sr.Report.SetRowHeight(sheetLabel, 1, 20)
	sr.Report.SetRowHeight(sheetLabel, 2, 20)
	sr.Report.SetCellStyle(sheetLabel, GetCellID(0, 1), GetCellID(0, 1), style)
	sr.Report.SetCellValue(sheetLabel, GetCellID(0, 1), startDate.Format("January"))

	style = sr.Styles["weekday"]
	sr.Report.SetCellStyle(sheetLabel, GetCellID(0, 2), GetCellID(0, 2), style)
	sr.Report.SetCellValue(sheetLabel, GetCellID(0, 2), sr.Date.Format("01/02/2006"))

	current := time.Date(month.Year(), month.Month(), 1, 0, 0, 0, 0, time.UTC)
	for current.Before(endDate) {
		styleID := "weekday"
		if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
			styleID = "weekend"
		}
		style = sr.Styles[styleID]
		sr.Report.SetCellStyle(sheetLabel, GetCellID(current.Day(), 1),
			GetCellID(current.Day(), 2), style)
		weekday := current.Format("Mon")
		sr.Report.SetCellValue(sheetLabel, GetCellID(current.Day(), 1), weekday[0:2])
		sr.Report.SetCellValue(sheetLabel, GetCellID(current.Day(), 2),
			strconv.Itoa(current.Day()))
		current = current.AddDate(0, 0, 1)
	}

	row := 2
	for _, wc := range sr.Workcenters {
		row++
		style = sr.Styles["workcenter"]
		sr.Report.SetRowHeight(sheetLabel, row, 20)
		sr.Report.SetCellStyle(sheetLabel, GetCellID(0, row),
			endColumn+strconv.Itoa(row), style)
		sr.Report.MergeCell(sheetLabel, GetCellID(0, row),
			endColumn+strconv.Itoa(row))
		sr.Report.SetCellValue(sheetLabel, GetCellID(0, row), wc.Name)
		sort.Sort(sites.ByPosition(wc.Positions))
		sort.Sort(sites.ByShift(wc.Shifts))
		for _, pos := range wc.Positions {
			sort.Sort(employees.ByEmployees(pos.Employees))
			for _, emp := range pos.Employees {
				row++
				sr.CreateEmployeeRow(sheetLabel, startDate, endDate, row, &emp)
			}
		}
		for _, sft := range wc.Shifts {
			sort.Sort(employees.ByEmployees(sft.Employees))
			for _, emp := range sft.Employees {
				row++
				sr.CreateEmployeeRow(sheetLabel, startDate, endDate, row, &emp)
			}
		}
	}
	return nil
}

func (sr *SiteScheduleReport) CreateEmployeeRow(sheetLabel string,
	start, end time.Time, row int, emp *employees.Employee) {
	styleID := "weekday"
	if row%2 == 0 {
		styleID = "evenday"
	}
	lastWorked := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, wc := range sr.Workcenters {
		if len(wc.Positions) > 0 {
			for _, pos := range wc.Positions {
				if len(pos.Employees) > 0 {
					for _, e := range pos.Employees {
						if strings.EqualFold(e.CompanyInfo.Company, emp.CompanyInfo.Company) &&
							len(e.Work) > 0 {
							for _, wk := range e.Work {
								if wk.DateWorked.After(lastWorked) {
									lastWorked = wk.DateWorked
								}
							}
						}
					}
				}
			}
		}
		if len(wc.Shifts) > 0 {
			for _, pos := range wc.Shifts {
				if len(pos.Employees) > 0 {
					for _, e := range pos.Employees {
						if strings.EqualFold(e.CompanyInfo.Company, emp.CompanyInfo.Company) &&
							len(e.Work) > 0 {
							for _, wk := range e.Work {
								if wk.DateWorked.After(lastWorked) {
									lastWorked = wk.DateWorked
								}
							}
						}
					}
				}
			}
		}
	}
	style := sr.Styles[styleID]
	sr.Report.SetRowHeight(sheetLabel, row, 20)
	sr.Report.SetCellStyle(sheetLabel, GetCellID(0, row), GetCellID(0, row), style)
	sr.Report.SetCellValue(sheetLabel, GetCellID(0, row),
		emp.Name.LastName+", "+emp.Name.FirstName[0:1])

	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
		time.UTC)
	for current.Before(end) {
		wd := emp.GetWorkday(current, lastWorked)
		code := ""
		styleID = "weekday"
		if wd != nil && wd.Code != "" {
			code = wd.Code
			if !sr.Workcodes[wd.Code] {
				styleID = wd.Code
			}
		}
		if styleID == "weekday" {
			if current.Weekday() == time.Saturday || current.Weekday() == time.Sunday {
				if row%2 == 0 {
					styleID = "evenend"
				} else {
					styleID = "weekend"
				}
			} else {
				if row%2 == 0 {
					styleID = "evenday"
				}
			}
		}
		style = sr.Styles[styleID]
		cellID := GetCellID(current.Day(), row)
		sr.Report.SetCellStyle(sheetLabel, cellID, cellID, style)
		sr.Report.SetCellValue(sheetLabel, cellID, code)
		current = current.AddDate(0, 0, 1)
	}
}

func (sr *SiteScheduleReport) CreateLegendSheet() error {
	sheetLabel := "Legend"
	sr.Report.NewSheet(sheetLabel)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	sr.Report.SetSheetView(sheetLabel, 0, &options)
	sr.Report.SetColWidth(sheetLabel, "A", "A", 30)

	team, err := svcs.GetTeam(sr.TeamID)
	if err != nil {
		return err
	}

	sort.Sort(labor.ByWorkcode(team.Workcodes))
	row := 0
	for _, wc := range team.Workcodes {
		if !strings.EqualFold(wc.BackColor, "ffffff") {
			row++
			sr.Report.SetRowHeight(sheetLabel, row, 20)
			style := sr.Styles[wc.Id]
			sr.Report.SetCellStyle(sheetLabel, GetCellID(0, row), GetCellID(0, row), style)
			sr.Report.SetCellValue(sheetLabel, GetCellID(0, row), wc.Title)
		}
	}
	return nil
}
