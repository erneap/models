package reports

import (
	"sort"
	"strconv"
	"time"

	"github.com/erneap/go-models/employees"
	"github.com/erneap/go-models/svcs"
	"github.com/xuri/excelize/v2"
)

type MidShift struct {
	Name employees.EmployeeName
	Mid  employees.Variation
}

type ByMidShifts []MidShift

func (c ByMidShifts) Len() int { return len(c) }
func (c ByMidShifts) Less(i, j int) bool {
	if c[i].Mid.StartDate.Equal(c[j].Mid.StartDate) {
		if c[i].Mid.EndDate.Equal(c[j].Mid.EndDate) {
			if c[i].Name.LastName == c[j].Name.LastName {
				if c[i].Name.FirstName == c[j].Name.FirstName {
					return c[i].Name.MiddleName < c[j].Name.MiddleName
				}
				return c[i].Name.FirstName < c[j].Name.FirstName
			}
			return c[i].Name.LastName < c[j].Name.LastName
		}
		return c[i].Mid.EndDate.Before(c[j].Mid.EndDate)
	}
	return c[i].Mid.StartDate.Before(c[j].Mid.StartDate)
}
func (c ByMidShifts) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (c *MidShift) GetDaysOff() string {
	weekdays := [7]bool{false, false, false, false, false,
		false, false}
	for i := 0; i < len(c.Mid.Schedule.Workdays); i++ {
		day := i % 7
		if c.Mid.Schedule.Workdays[i].Code != "" {
			weekdays[day] = true
		}
	}

	days := [7]string{"Su", "Mo", "Tu", "We", "Th", "Fr", "Sa"}
	answer := ""
	for d, dy := range weekdays {
		if !dy {
			if answer != "" {
				answer += ","
			}
			answer += days[d]
		}
	}
	return answer
}

type MidShiftReport struct {
	Report      *excelize.File
	Date        time.Time
	TeamID      string
	SiteID      string
	Styles      map[string]int
	Employees   []employees.Employee
	MidShifts   []MidShift
	CurrentAsOf time.Time
}

func (m *MidShiftReport) Create() error {
	m.CurrentAsOf = time.Now().UTC()
	m.Styles = make(map[string]int)
	m.Report = excelize.NewFile()

	site, err := svcs.GetSite(m.TeamID, m.SiteID)
	if err != nil {
		return err
	}
	m.Employees = append(m.Employees, site.Employees...)

	m.Date = time.Date(m.Date.Year(), time.Month(1), 1, 0,
		0, 0, 0, time.UTC)
	for _, emp := range m.Employees {
		for _, vari := range emp.Data.Variations {
			if vari.IsMids {
				if vari.StartDate.After(m.Date) ||
					vari.EndDate.After(m.Date) {
					mid := MidShift{
						Name: emp.Name,
						Mid:  vari,
					}
					m.MidShifts = append(m.MidShifts, mid)
				}
			}
		}
	}

	err = m.SetStyles()
	if err != nil {
		return nil
	}

	sort.Sort(ByMidShifts(m.MidShifts))
	year := 0
	row := 0
	sheetName := ""
	for v, vari := range m.MidShifts {
		mYear := vari.Mid.StartDate.Year()
		if vari.Mid.StartDate.Before(m.Date) {
			mYear = vari.Mid.EndDate.Year()
		}
		if mYear != year {
			year = mYear
			sheetName = strconv.Itoa(year)
			m.Report.NewSheet(sheetName)
			options := excelize.ViewOptions{}
			options.ShowGridLines = &[]bool{false}[0]
			m.Report.SetSheetView(sheetName, 0, &options)

			// set column widths
			m.Report.SetColWidth(sheetName, "A", "A", 30.0)
			m.Report.SetColWidth(sheetName, "B", "D", 10.0)

			// add sheet header in row 1
			label := sheetName + " MIDS ROTATION SCHEDULE"
			style := m.Styles["header"]
			m.Report.SetCellStyle(sheetName, "A1", "D1", style)
			m.Report.MergeCell(sheetName, "A1", "D1")
			m.Report.SetCellValue(sheetName, "A1", label)

			// add sub headers in row 2
			style = m.Styles["subheader"]
			m.Report.SetCellStyle(sheetName, "A2", "D2", style)
			m.Report.SetCellValue(sheetName, "A2", "NAME")
			m.Report.SetCellValue(sheetName, "B2", "START")
			m.Report.SetCellValue(sheetName, "C2", "END")
			m.Report.SetCellValue(sheetName, "D2", "DAYS OFF")
			row = 2
		}
		row++
		style := m.Styles["even"]
		if v%2 == 1 {
			style = m.Styles["odd"]
		}
		cellID1 := GetCellID(0, row)
		cellID2 := GetCellID(3, row)
		m.Report.SetCellStyle(sheetName, cellID1, cellID2, style)
		m.Report.SetCellValue(sheetName, cellID1,
			vari.Name.GetLastFirst())
		cellID1 = GetCellID(1, row)
		m.Report.SetCellValue(sheetName, cellID1,
			vari.Mid.StartDate.Format("01/02/2006"))
		cellID1 = GetCellID(2, row)
		m.Report.SetCellValue(sheetName, cellID1,
			vari.Mid.EndDate.Format("01/02/2006"))
		cellID1 = GetCellID(3, row)
		m.Report.SetCellValue(sheetName, cellID1,
			vari.GetDaysOff())
	}

	m.Report.DeleteSheet("Sheet1")

	return nil
}

func (m *MidShiftReport) SetStyles() error {
	style, err := m.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "ffffff", Style: 2},
			{Type: "top", Color: "ffffff", Style: 2},
			{Type: "right", Color: "ffffff", Style: 2},
			{Type: "bottom", Color: "ffffff", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"0066cc"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "ffffff", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	m.Styles["header"] = style
	style, err = m.Report.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "ffffff", Style: 2},
			{Type: "top", Color: "ffffff", Style: 2},
			{Type: "right", Color: "ffffff", Style: 2},
			{Type: "bottom", Color: "ffffff", Style: 2},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 10, Color: "ffffff", Family: "Calibri Light"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	m.Styles["subheader"] = style
	style, err = m.Report.NewStyle(&excelize.Style{
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
	m.Styles["even"] = style
	style, err = m.Report.NewStyle(&excelize.Style{
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
	m.Styles["odd"] = style

	return nil
}
