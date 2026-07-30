package reports

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/models/v2/config"
	"github.com/erneap/models/v2/metrics"
	"github.com/erneap/models/v2/systemdata"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type DrawSummary struct {
	ReportType   systemdata.GeneralTypes
	ReportPeriod uint
	StartDate    time.Time
	EndDate      time.Time
	Daily        bool
	Missions     []MissionDay
	Outages      []OutageDay
	Styles       map[string]int
	SystemInfo   systemdata.SystemInfo
}

func (ds *DrawSummary) Create() (*excelize.File, error) {

	// create outage excel file
	workbook := excelize.NewFile()
	ds.SystemInfo = metrics.InitialData()

	switch ds.ReportPeriod {
	case 1:
		ds.Daily = false
		ds.EndDate = ds.StartDate.Add(24 * time.Hour).Add(-1 * time.Second)
	case 7:
		ds.EndDate = ds.StartDate.Add(7 * 24 * time.Hour).Add(-1 * time.Second)
	case 30:
		ds.StartDate = time.Date(ds.StartDate.Year(), ds.StartDate.Month(), 1, 0, 0,
			0, 0, time.UTC)
		ds.EndDate = ds.StartDate.AddDate(0, 1, 0).Add(-1 * time.Second)
		ds.Daily = false
	case 365:
		ds.StartDate = time.Date(ds.StartDate.Year(), time.January, 1, 0, 0, 0, 0,
			time.UTC)
		ds.EndDate = ds.StartDate.AddDate(1, 0, 0).Add(-1 * time.Second)
		ds.Daily = false
	}

	// create mission and outage day arrays
	dayStart := time.Date(ds.StartDate.Year(), ds.StartDate.Month(),
		ds.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	for tm := dayStart; tm.Before(ds.EndDate) || tm.Equal(ds.EndDate); tm = tm.AddDate(0, 0, 1) {
		mday := MissionDay{
			MissionDate: tm,
		}
		ds.Missions = append(ds.Missions, mday)
		oday := OutageDay{
			OutageDate: tm,
		}
		ds.Outages = append(ds.Outages, oday)
	}
	sort.Sort(ByMissionDay(ds.Missions))

	// get missions for the time period and fill them into the mission days
	var tmissions []metrics.Mission
	filter := bson.M{"missionDate": bson.M{"$gte": ds.StartDate, "$lte": ds.EndDate}}
	cursor, err := config.GetCollection(config.DB, "metrics", "missions").Find(context.TODO(),
		filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &tmissions); err != nil {
		return nil, err
	}

	for _, msn := range tmissions {
		for pos, mday := range ds.Missions {
			if mday.MissionDate.Equal(msn.MissionDate) {
				mday.Missions = append(mday.Missions, msn)
				ds.Missions[pos] = mday
			}
		}
	}
	// get outages for the time period and fill them into the outage days
	var tOutages []metrics.GroundOutage
	filter = bson.M{"outageDate": bson.M{"$gte": ds.StartDate, "$lte": ds.EndDate}}
	cursor, err = config.GetCollection(config.DB, "metrics", "groundoutages").Find(context.TODO(),
		filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &tOutages); err != nil {
		return nil, err
	}

	for _, outage := range tOutages {
		for pos, oday := range ds.Outages {
			if oday.OutageDate.Equal(outage.OutageDate) && outage.Capability == "NMC" {
				oday.Outages = append(oday.Outages, outage)
				ds.Outages[pos] = oday
			}
		}
	}

	err = ds.CreateWorkbookStyles(workbook)
	if err != nil {
		return nil, err
	}
	err = ds.CreateMissionSheet(workbook)
	if err != nil {
		return nil, err
	}

	err = ds.CreateOutageSheet(workbook)
	if err != nil {
		return nil, err
	}
	workbook.DeleteSheet("Sheet1")
	return workbook, nil
}

func (ds *DrawSummary) CreateWorkbookStyles(workbook *excelize.File) error {
	ds.Styles = make(map[string]int)
	style, err := workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 14, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawMissionTitle"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawMissionDate"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawMissionData"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawMissionTime"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"000000"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutageTitle"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesDateEven"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D3D3D3"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesDateOdd"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawNoOutagesEven"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D3D3D3"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawNoOutagesOdd"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesDataEven"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D3D3D3"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesDataOdd"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesProblemEven"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"D3D3D3"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 10, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ds.Styles["drawOutagesProblemOdd"] = style
	return nil
}

func (ds *DrawSummary) CreateMissionSheet(workbook *excelize.File) error {
	// add new sheet with label of Draw Missions
	label := "DRAW Missions"
	workbook.NewSheet(label)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	workbook.SetSheetView(label, 0, &options)
	// set column widths with all columns at 8.43 except column B at 75.0
	err := workbook.SetColWidth(label, "A", "J", float64(8.43))
	if err != nil {
		return err
	}
	err = workbook.SetColWidth(label, "B", "B", float64(75.0))
	if err != nil {
		return err
	}

	// add label
	style := ds.Styles["drawMissionTitle"]
	err = workbook.SetCellStyle(label, "B2", "B2", style)
	if err != nil {
		return err
	}
	err = workbook.SetCellValue(label, "B2", "DRAW Mission Summary")
	if err != nil {
		return err
	}
	nRow := 2

	// loop through the mission days to load each mission's data
	for _, msnday := range ds.Missions {
		nRow += 2

		style = ds.Styles["drawMissionDate"]
		err = workbook.SetCellStyle(label, ds.GetCellID("B", nRow),
			ds.GetCellID("B", nRow), style)
		if err != nil {
			return err
		}
		err = workbook.SetCellValue(label, ds.GetCellID("B", nRow),
			ds.GetDateString(msnday.MissionDate))
		if err != nil {
			return err
		}
		if len(msnday.Missions) == 0 {
			style = ds.Styles["drawMissionData"]
			err = workbook.SetCellStyle(label, ds.GetCellID("B", nRow+1),
				ds.GetCellID("B", nRow+1), style)
			if err != nil {
				return err
			}
			err = workbook.SetCellValue(label, ds.GetCellID("B", nRow+1),
				"No Missions schedule on this date")
			if err != nil {
				return err
			}
			nRow++
		} else {
			sort.Sort(metrics.ByMission(msnday.Missions))
			for msns, msn := range msnday.Missions {

				showMission := ""
				u2Text := ""
				sensorOutage := uint(0)
				groundOutage := uint(0)
				partialHB := -1
				partialLB := -1
				sensorComments := msn.Comments
				executeMinutes := uint(0)
				for _, plat := range ds.SystemInfo.Platforms {
					if strings.EqualFold(plat.ID, msn.PlatformID) {
						for _, mSen := range msn.Sensors {
							for _, pSen := range plat.Sensors {
								if strings.EqualFold(mSen.SensorID, pSen.ID) &&
									pSen.UseForExploitation(msn.Exploitation, ds.ReportType) {
									sensorOutage = mSen.SensorOutage.TotalOutageMinutes
									groundOutage = mSen.GroundOutage
									sensorComments = strings.TrimSpace(sensorComments)
									if sensorComments != "" {
										sensorComments += ", "
									}
									sensorComments += mSen.Comments
									executeMinutes = mSen.ExecutedMinutes + mSen.AdditionalMinutes
									switch pSen.ID {
									case "PME3", "PME4":
										showMission = "- CWS/" + pSen.ID + ": Kit " +
											mSen.KitNumber + " was Code " +
											strconv.FormatUint(uint64(mSen.FinalCode), 10)
										u2Text = " with " + pSen.Association + " and " +
											msn.Communications
									case "PME9":
										showMission = "- DDSA/PME9: Kit " +
											mSen.KitNumber + " was Code " +
											strconv.FormatUint(uint64(mSen.FinalCode), 10)
									case "PME12":
										showMission = "- CICS/PME12: Kit " +
											mSen.KitNumber + " was Code " +
											strconv.FormatUint(uint64(mSen.FinalCode), 10)
										partialHB = int(mSen.SensorOutage.PartialHBOutageMinutes)
										partialLB = int(mSen.SensorOutage.PartialLBOutageMinutes)
									case "IMINT":
										showMission = "- CWS/IMINT Sensor"
									}
								}
							}
						}
					}
				}
				if showMission != "" {
					msns++
					text := "This mission on " + ds.GetDateString(msn.MissionDate) +
						" was " + strings.ToUpper(msn.PlatformID)
					if msn.TailNumber != "" {
						text += " using Article " + msn.TailNumber
					}
					if u2Text != "" {
						text += u2Text
					}
					if !strings.EqualFold(msn.Exploitation, "primary") {
						text += "\n\r- Primary exploitating DGS was DGS-" +
							msn.PrimaryDCGS
					}
					if !msn.Cancelled && !msn.Aborted &&
						!msn.IndefDelay {
						text += "\n\r" + showMission
					}
					if msn.Aborted {
						text += "\n\r- Mission Aborted early"
					}
					if msn.Cancelled {
						text += "\n\r- Mission Cancelled"
					}
					if msn.IndefDelay {
						text += "\n\r- Mission Indefinite Delay"
					}
					if strings.EqualFold(msn.Exploitation, "primary") &&
						!msn.Aborted && !msn.Cancelled &&
						!msn.IndefDelay {
						text += "\n\r- Ground Outages: " +
							strconv.FormatUint(uint64(groundOutage), 10) + " mins."
						text += "\n\r- Sensor Outages: " +
							strconv.FormatUint(uint64(sensorOutage), 10) + " mins."
						if partialHB >= 0 {
							text += "\n\r- Partial HB Outages: " +
								strconv.FormatInt(int64(partialHB), 10) + " mins."
						}
						if partialLB >= 0 {
							text += "\n\r- Partial LB Outages: " +
								strconv.FormatInt(int64(partialLB), 10) + " mins."
						}
					}
					if sensorComments != "" {
						text += "\n\r- Comments: " + sensorComments
					}
					for text[len(text)-1:] == "\n" || text[len(text)-1:] == "\r" {
						text = text[:len(text)-1]
					}
					textrows := len(strings.Split(text, "\n\r"))
					workbook.SetRowHeight(label, nRow+msns, float64(textrows)*13.0)

					style = ds.Styles["drawMissionData"]
					workbook.SetCellStyle(label, ds.GetCellID("B", nRow+msns),
						ds.GetCellID("B", nRow+msns), style)
					workbook.SetCellValue(label, ds.GetCellID("B", nRow+msns), text)

					style = ds.Styles["drawMissionTime"]
					workbook.SetCellStyle(label, ds.GetCellID("C", nRow+msns),
						ds.GetCellID("D", nRow+msns), style)
					workbook.SetCellValue(label, ds.GetCellID("C", nRow+msns),
						ds.GetTimeString(executeMinutes))
					workbook.SetCellValue(label, ds.GetCellID("D", nRow+msns),
						ds.GetTimeString(executeMinutes-(sensorOutage+groundOutage)))
				}
			}
			nRow += len(msnday.Missions) - 1
		}
	}
	return nil
}

func (ds *DrawSummary) CreateOutageSheet(workbook *excelize.File) error {
	label := "DRAW Outage"

	workbook.NewSheet(label)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	workbook.SetSheetView(label, 0, &options)
	// set the column widths
	workbook.SetColWidth(label, "A", "J", 8.43)
	workbook.SetColWidth(label, "B", "B", 11.0)
	workbook.SetColWidth(label, "C", "C", 11.0)
	workbook.SetColWidth(label, "D", "D", 12.0)
	workbook.SetColWidth(label, "E", "E", 8.0)
	workbook.SetColWidth(label, "F", "F", 110.0)

	workbook.SetRowHeight(label, 2, 32.0)

	nRow := 2
	style := ds.Styles["drawOutageTitle"]
	workbook.SetCellStyle(label, ds.GetCellID("B", nRow), ds.GetCellID("F", nRow),
		style)

	workbook.SetCellValue(label, ds.GetCellID("B", nRow), "DATE")
	workbook.SetCellValue(label, ds.GetCellID("C", nRow), "SYSTEM")
	workbook.SetCellValue(label, ds.GetCellID("D", nRow), "SUBSYSTEM")
	workbook.SetCellValue(label, ds.GetCellID("E", nRow), "Outage (Mins)")
	workbook.SetCellValue(label, ds.GetCellID("F", nRow),
		"PROBLEM(S)/RESOLUTION(S)")
	nRow++

	for _, outDay := range ds.Outages {
		styleEnd := "Even"
		if nRow%2 == 1 {
			styleEnd = "Odd"
		}
		style = ds.Styles["drawOutagesDate"+styleEnd]
		workbook.SetCellStyle(label, ds.GetCellID("B", nRow),
			ds.GetCellID("B", nRow), style)
		workbook.SetCellValue(label, ds.GetCellID("B", nRow),
			ds.GetDateString(outDay.OutageDate))
		if len(outDay.Outages) == 0 {
			style = ds.Styles["drawNoOutages"+styleEnd]
			workbook.SetCellStyle(label, ds.GetCellID("C", nRow),
				ds.GetCellID("F", nRow), style)
			workbook.MergeCell(label, ds.GetCellID("C", nRow), ds.GetCellID("F", nRow))
			workbook.SetCellValue(label, ds.GetCellID("C", nRow), "NSTR")
		} else {
			for nOut, outage := range outDay.Outages {
				styleEnd = "Even"
				if (nRow+nOut)%2 == 1 {
					styleEnd = "Odd"
				}
				if nOut > 0 {
					style = ds.Styles["drawOutagesDate"+styleEnd]
					workbook.SetCellStyle(label, ds.GetCellID("B", nRow+nOut),
						ds.GetCellID("B", nRow+nOut), style)
					workbook.SetCellValue(label, ds.GetCellID("B", nRow+nOut), "")
				}
				style = ds.Styles["drawOutagesData"+styleEnd]
				workbook.SetCellStyle(label, ds.GetCellID("C", nRow+nOut),
					ds.GetCellID("E", nRow+nOut), style)
				workbook.SetCellValue(label, ds.GetCellID("C", nRow+nOut),
					strings.ToUpper(outage.GroundSystem))
				workbook.SetCellValue(label, ds.GetCellID("D", nRow+nOut),
					strings.ToUpper(outage.Subsystem))
				workbook.SetCellValue(label, ds.GetCellID("E", nRow+nOut),
					strconv.FormatUint(uint64(outage.OutageMinutes), 10))

				text := "PROBLEM(s): " + outage.Problem + "\rRESOLUTION(s)" +
					outage.FixAction
				style = ds.Styles["drawOutagesProblem"+styleEnd]
				workbook.SetCellStyle(label, ds.GetCellID("F", nRow+nOut),
					ds.GetCellID("F", nRow+nOut), style)
				workbook.SetCellValue(label, ds.GetCellID("F", nRow+nOut), text)
				workbook.SetRowHeight(label, nRow+nOut, 26.0)
			}
			nRow += len(outDay.Outages) - 1
		}
		nRow++
	}
	return nil
}

func (ds *DrawSummary) GetCellID(letter string, row int) string {
	return letter + strconv.FormatInt(int64(row), 10)
}

func (ds *DrawSummary) GetDateString(date time.Time) string {
	return date.Format("02 Jan 2006")
}

func (ds *DrawSummary) GetTimeString(minutes uint) string {
	hours := minutes / 60
	mins := minutes - (hours * 60)
	answer := ""
	if hours < 10 {
		answer += "0"
	}
	answer += strconv.FormatInt(int64(hours), 10) + ":"
	if mins < 10 {
		answer += "0"
	}
	answer += strconv.FormatInt(int64(mins), 10)
	return answer
}
