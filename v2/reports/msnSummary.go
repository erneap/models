package reports

import (
	"context"
	"fmt"
	"log"
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

type MissionSummary struct {
	ReportType   systemdata.GeneralTypes
	ReportPeriod uint
	StartDate    time.Time
	EndDate      time.Time
	Daily        bool
	Missions     []metrics.Mission
	Outages      []metrics.GroundOutage
	Styles       map[string]int
	SystemInfo   systemdata.SystemInfo
}

func (ms *MissionSummary) Create() (*excelize.File, error) {
	// create outage excel file
	workbook := excelize.NewFile()
	ms.SystemInfo = metrics.InitialData()

	switch ms.ReportPeriod {
	case 1:
		ms.Daily = false
		ms.EndDate = ms.StartDate.Add(24 * time.Hour).Add(-1 * time.Second)
	case 7:
		ms.EndDate = ms.StartDate.Add(7 * 24 * time.Hour).Add(-1 * time.Second)
	case 30:
		ms.StartDate = time.Date(ms.StartDate.Year(), ms.StartDate.Month(), 1, 0, 0,
			0, 0, time.UTC)
		ms.EndDate = ms.StartDate.AddDate(0, 1, 0).Add(-1 * time.Second)
		ms.Daily = false
	case 365:
		ms.StartDate = time.Date(ms.StartDate.Year(), time.January, 1, 0, 0, 0, 0,
			time.UTC)
		ms.EndDate = ms.StartDate.AddDate(1, 0, 0).Add(-1 * time.Second)
		ms.Daily = false
	}

	// collect all the missions for the period
	var tmissions []metrics.Mission
	filter := bson.M{"missionDate": bson.M{"$gte": ms.StartDate, "$lte": ms.EndDate}}
	cursor, err := config.GetCollection(config.DB, "metrics", "missions").Find(context.TODO(),
		filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &tmissions); err != nil {
		return nil, err
	}

	ms.Missions = append(ms.Missions, tmissions...)

	sort.Sort(metrics.ByMission(ms.Missions))
	log.Println(len(ms.Missions))

	// collect all the outages for the period
	var tOutages []metrics.GroundOutage
	filter = bson.M{"outageDate": bson.M{"$gte": ms.StartDate, "$lte": ms.EndDate}}
	cursor, err = config.GetCollection(config.DB, "metrics", "groundoutages").Find(context.TODO(),
		filter)
	if err != nil {
		return nil, err
	}

	if err = cursor.All(context.TODO(), &tOutages); err != nil {
		return nil, err
	}

	for _, outage := range tOutages {
		outage.Decrypt()
		ms.Outages = append(ms.Outages, outage)
	}

	sort.Sort(metrics.ByOutage(ms.Outages))

	// create the report sheets based on ReportType, and the rest of the information
	// but create the needed styles first
	err = ms.CreateSummaryStyles(workbook)
	if err != nil {
		return nil, err
	}
	ms.AddSummarySheet(workbook, ms.StartDate, ms.EndDate, ms.Daily,
		ms.ReportType, "")

	// remove the default sheet from the workbook
	workbook.DeleteSheet("Sheet1")
	return workbook, nil
}

func (ms *MissionSummary) CreateSummaryStyles(workbook *excelize.File) error {
	ms.Styles = make(map[string]int)
	style, err := workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"CCCCCC"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 12, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["labelHeader"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"B8CCE4"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformHeader"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E5B8B7"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformLabelLeft"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"E5B8B7"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformLabelCenter"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "right", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorRight"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorData"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		NumFmt: 2,
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorDataNum"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 9, Color: "0070c0"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorDataNotNorm"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: false, Size: 9, Color: "0070c0"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
		NumFmt: 2,
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorDataNumNotNorm"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["platformSensorCenter"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 0},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["gsSystemIDOnly"] = style
	style, err = workbook.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 0},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"FFFFFF"}, Pattern: 1},
		Font: &excelize.Font{Bold: true, Size: 9, Color: "000000"},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center",
			WrapText: true},
	})
	if err != nil {
		return err
	}
	ms.Styles["gsEnclaveNone"] = style
	return nil
}

func (ms *MissionSummary) AddSummarySheet(workbook *excelize.File,
	start, end time.Time, daily bool, rptType systemdata.GeneralTypes,
	label string) {
	exploitations := []string{"Primary", "Shadow/Federated"}

	// pull the missions from the period and map them by platform and exploitation
	missionMap := make(map[string]MissionType)
	for _, msn := range ms.Missions {
		if (msn.MissionDate.Equal(start) || msn.MissionDate.After(start)) &&
			(msn.MissionDate.Equal(end) || msn.MissionDate.Before(end)) {
			exp := msn.Exploitation
			if !strings.EqualFold(exp, "primary") {
				exp = "Shadow/Federated"
			}
			key := strings.ToUpper(exp) + "-" + msn.PlatformID
			mType, ok := missionMap[key]
			if !ok {
				mType = MissionType{
					Exploitation: exp,
					Platform:     msn.PlatformID,
				}
			}
			mType.Missions = append(mType.Missions, msn)
			missionMap[key] = mType
		}
	}

	// compile composite sensor list
	var sensors []string
	for _, mType := range missionMap {
		sens := mType.GetSensorList()
		for _, sen := range sens {
			found := false
			for _, s := range sensors {
				if strings.EqualFold(sen, s) {
					found = true
				}
			}
			if !found {
				sensors = append(sensors, sen)
			}
		}
	}

	// pull ground outages for period
	var outages []metrics.GroundOutage
	for _, outage := range ms.Outages {
		if (outage.OutageDate.Equal(start) || outage.OutageDate.After(start)) &&
			(outage.OutageDate.Equal(end) || outage.OutageDate.Before(end) &&
				outage.Capability == "NMC") {
			outages = append(outages, outage)
		}
	}

	// create worksheet for this summary
	if label == "" {
		label = "Summary"
		if end.Sub(start).Hours() <= 24 {
			label = strconv.FormatInt(int64(start.Day()), 10) + " " + start.Month().String()[:3]
		}
	}
	workbook.NewSheet(label)
	options := excelize.ViewOptions{}
	options.ShowGridLines = &[]bool{false}[0]
	workbook.SetSheetView(label, 0, &options)

	// set the sheet column widths
	workbook.SetColWidth(label, "A", "L", 8)
	workbook.SetColWidth(label, "C", "C", 4)
	workbook.SetColWidth(label, "B", "B", 7)
	workbook.SetColWidth(label, "D", "D", 7)

	// create headers/labels at top
	workbook.SetRowHeight(label, 3, 18)
	workbook.MergeCell(label, "B3", "D3")
	style := ms.Styles["labelHeader"]
	workbook.SetCellStyle(label, "B3", "J4", style)
	workbook.SetCellValue(label, "B3", "Summary Period")
	workbook.MergeCell(label, "E3", "F3")
	workbook.SetCellValue(label, "E3", "From:")
	workbook.MergeCell(label, "G3", "H3")
	workbook.SetCellValue(label, "G3", "To:")
	workbook.MergeCell(label, "I3", "J3")
	workbook.SetCellValue(label, "I3", "Execution Time")
	workbook.MergeCell(label, "B4", "D4")
	workbook.SetCellValue(label, "B4", "")
	workbook.MergeCell(label, "E4", "F4")
	workbook.SetCellValue(label, "E4", ms.GetDateString(start))
	workbook.MergeCell(label, "G4", "H4")
	workbook.SetCellValue(label, "G4", ms.GetDateString(end))
	execution := uint(0)
	for _, mType := range missionMap {
		execution += (mType.GetExecutedTime(sensors, nil) +
			mType.GetAdditional(sensors, nil)) - mType.GetOverlap()
	}
	workbook.MergeCell(label, "I4", "J4")
	workbook.SetCellValue(label, "I4", ms.GetTimeString(execution))

	// summary of exploitations/platforms and sensors
	style = ms.Styles["platformHeader"]
	workbook.SetCellStyle(label, "B6", "J7", style)
	workbook.MergeCell(label, "B6", "F7")
	workbook.SetCellValue(label, "B6", "SYSTEM")
	workbook.MergeCell(label, "G6", "J6")
	workbook.SetCellValue(label, "G6", "SORTIES")
	workbook.SetCellValue(label, "G7", "SCHED")
	workbook.SetCellValue(label, "H7", "EXEC")
	workbook.SetCellValue(label, "I7", "CANCEL")
	workbook.SetCellValue(label, "J7", "ABORT")

	nRow := 7
	for _, exp := range exploitations {
		var sensorList []string
		for _, plat := range ms.SystemInfo.Platforms {
			for _, pSen := range plat.Sensors {
				bUse := false
				for _, e := range pSen.Exploitations {
					if strings.EqualFold(e.Exploitation, exp) &&
						(rptType == systemdata.ALL ||
							(rptType == systemdata.GEOINT && e.ShowOnGEOINT) ||
							(rptType == systemdata.MIST && e.ShowOnMIST) ||
							(rptType == systemdata.SYERS && e.ShowOnGSEG) ||
							(rptType == systemdata.XINT && e.ShowOnXINT)) {
						bUse = true
					}
				}
				if bUse {
					found := false
					for _, s := range sensorList {
						if strings.EqualFold(pSen.ID, s) {
							found = true
						}
					}
					if !found {
						sensorList = append(sensorList, pSen.ID)
					}
				}
			}
		}
		nRow++
		style = ms.Styles["platformLabelLeft"]
		workbook.SetCellStyle(label, ms.GetCellID("B", nRow), ms.GetCellID("F", nRow),
			style)
		workbook.MergeCell(label, ms.GetCellID("B", nRow), ms.GetCellID("F", nRow))
		workbook.SetCellValue(label, ms.GetCellID("B", nRow), exp)
		style = ms.Styles["platformLabelCenter"]
		workbook.SetCellStyle(label, ms.GetCellID("G", nRow), ms.GetCellID("J", nRow),
			style)
		workbook.SetCellValue(label, ms.GetCellID("G", nRow),
			ms.GetTotalScheduled(exp, missionMap, sensorList))
		workbook.SetCellValue(label, ms.GetCellID("H", nRow),
			ms.GetTotalExecuted(exp, missionMap, sensorList))
		workbook.SetCellValue(label, ms.GetCellID("I", nRow),
			ms.GetTotalCancelled(exp, missionMap, sensorList))
		workbook.SetCellValue(label, ms.GetCellID("J", nRow),
			ms.GetTotalAborted(exp, missionMap, sensorList))

		styleRight := ms.Styles["platformSensorRight"]
		styleCenter := ms.Styles["platformSensorData"]
		for _, plat := range ms.SystemInfo.Platforms {
			key := strings.ToUpper(exp) + "-" + plat.ID
			mtype := missionMap[key]
			if plat.ShowOnSummary(exp, rptType) {
				nRow++
				platCount := -1
				workbook.MergeCell(label, ms.GetCellID("B", nRow), ms.GetCellID("C", nRow))
				workbook.SetCellStyle(label, ms.GetCellID("B", nRow),
					ms.GetCellID("C", nRow), styleRight)
				workbook.SetCellValue(label, ms.GetCellID("B", nRow), plat.ID)
				for _, pSen := range plat.Sensors {
					if pSen.UseForExploitation(exp, rptType) {
						platCount++
						if platCount <= 0 {
							workbook.MergeCell(label, ms.GetCellID("D", nRow+platCount),
								ms.GetCellID("F", nRow+platCount))
							workbook.SetCellStyle(label, ms.GetCellID("D", nRow+platCount),
								ms.GetCellID("F", nRow+platCount), styleRight)
							workbook.SetCellValue(label, ms.GetCellID("D", nRow+platCount),
								pSen.ID)
						} else {
							workbook.MergeCell(label, ms.GetCellID("B", nRow+platCount),
								ms.GetCellID("F", nRow+platCount))
							workbook.SetCellStyle(label, ms.GetCellID("B", nRow+platCount),
								ms.GetCellID("F", nRow+platCount), styleRight)
							workbook.SetCellValue(label, ms.GetCellID("B", nRow+platCount),
								pSen.ID)
						}
						workbook.SetCellStyle(label, ms.GetCellID("G", nRow+platCount),
							ms.GetCellID("J", nRow+platCount), styleCenter)
						workbook.SetCellValue(label, ms.GetCellID("G", nRow+platCount),
							mtype.GetScheduled(exp, []string{pSen.ID}))
						workbook.SetCellValue(label, ms.GetCellID("H", nRow+platCount),
							mtype.GetExecuted(exp, []string{pSen.ID}))
						workbook.SetCellValue(label, ms.GetCellID("I", nRow+platCount),
							mtype.GetCancelled(exp, []string{pSen.ID}))
						workbook.SetCellValue(label, ms.GetCellID("J", nRow+platCount),
							mtype.GetAborted(exp, []string{pSen.ID}))
					}
				}
				nRow += platCount
			}
		}
	}

	// Ground Systems usage matrix
	// Displayed ground systems are determined by report type
	nRow += 2
	style = ms.Styles["platformHeader"]
	workbook.SetCellStyle(label, ms.GetCellID("B", nRow),
		ms.GetCellID("K", nRow+1), style)
	workbook.MergeCell(label, ms.GetCellID("B", nRow), ms.GetCellID("D", nRow+1))
	workbook.SetCellValue(label, ms.GetCellID("B", nRow), "Ground System/Enclave")
	workbook.MergeCell(label, ms.GetCellID("E", nRow), ms.GetCellID("H", nRow))
	workbook.SetCellValue(label, ms.GetCellID("E", nRow), "HOURS")
	workbook.MergeCell(label, ms.GetCellID("I", nRow), ms.GetCellID("J", nRow))
	workbook.SetCellValue(label, ms.GetCellID("I", nRow), "Outages")
	workbook.MergeCell(label, ms.GetCellID("K", nRow), ms.GetCellID("K", nRow+1))
	workbook.SetCellValue(label, ms.GetCellID("K", nRow), "Ao%")
	nRow++
	workbook.SetCellValue(label, ms.GetCellID("E", nRow), "PRE")
	workbook.SetCellValue(label, ms.GetCellID("F", nRow), "PLAN")
	workbook.SetCellValue(label, ms.GetCellID("G", nRow), "EXEC")
	workbook.SetCellValue(label, ms.GetCellID("H", nRow), "POST")
	workbook.SetCellValue(label, ms.GetCellID("I", nRow), "#")
	workbook.SetCellValue(label, ms.GetCellID("J", nRow), "HOURS")

	for _, gs := range ms.SystemInfo.GroundSystems {
		if rptType == systemdata.ALL ||
			(rptType == systemdata.GEOINT && gs.ShowOnGEOINT) ||
			(rptType == systemdata.SYERS && gs.ShowOnGSEG) ||
			(rptType == systemdata.MIST && gs.ShowOnMIST) ||
			(rptType == systemdata.XINT && gs.ShowOnXINT) {
			premission := uint(0)
			scheduled := uint(0)
			executed := uint(0)
			postmission := uint(0)
			overlap := uint(0)
			outageNumber := uint(0)
			outageTime := uint(0)
			for _, mType := range missionMap {
				if len(gs.Exploitations) == 0 {
					for _, exp := range exploitations {
						if strings.EqualFold(exp, mType.Exploitation) {
							premission += mType.GetPremissionTime(sensors, nil)
							scheduled += mType.GetScheduledTime(sensors, nil)
							executed += (mType.GetExecutedTime(sensors, nil) +
								mType.GetAdditional(sensors, nil))
							overlap += mType.GetOverlap()
							postmission += mType.GetPostmissionTime(sensors, nil)
						}
					}
				} else {
					found := false
					for _, exp := range gs.Exploitations {
						if strings.EqualFold(exp.Exploitation, mType.Exploitation) &&
							strings.EqualFold(exp.PlatformID, mType.Platform) {
							found = true
						}
					}
					if found {
						premission += mType.GetPremissionTime(sensors, &gs)
						scheduled += mType.GetScheduledTime(sensors, &gs)
						executed += (mType.GetExecutedTime(sensors, &gs) +
							mType.GetAdditional(sensors, &gs))
						overlap += mType.GetOverlap()
						postmission += mType.GetPostmissionTime(sensors, &gs)
					}
				}
			}
			if executed >= overlap && overlap > 0 {
				executed -= overlap
			} else if executed < overlap {
				executed = uint(0)
			}
			encCount := -1
			nRow++
			for _, enclave := range gs.Enclaves {
				encCount++
				outageNumber = 0
				outageTime = 0
				if len(gs.Enclaves) == 1 {
					style = ms.Styles["gsSystemIDOnly"]
					workbook.SetCellStyle(label, ms.GetCellID("B", nRow),
						ms.GetCellID("C", nRow), style)
					workbook.MergeCell(label, ms.GetCellID("B", nRow),
						ms.GetCellID("C", nRow))
					workbook.SetCellValue(label, ms.GetCellID("B", nRow), gs.ID)
					style = ms.Styles["gsEnclaveNone"]
					workbook.SetCellStyle(label, ms.GetCellID("D", nRow),
						ms.GetCellID("D", nRow), style)
					workbook.SetCellValue(label, ms.GetCellID("D", nRow), "")
				} else {
					if encCount == 0 {
						style = ms.Styles["platformSensorCenter"]
						workbook.SetCellStyle(label, ms.GetCellID("B", nRow),
							ms.GetCellID("C", nRow), style)
						workbook.MergeCell(label, ms.GetCellID("B", nRow),
							ms.GetCellID("C", nRow))
						workbook.SetCellValue(label, ms.GetCellID("B", nRow),
							gs.ID)
						style = ms.Styles["platformSensorRight"]
						workbook.SetCellStyle(label, ms.GetCellID("D", nRow),
							ms.GetCellID("D", nRow), style)
						workbook.SetCellValue(label, ms.GetCellID("D", nRow),
							enclave)
					} else {
						style = ms.Styles["platformSensorRight"]
						workbook.SetCellStyle(label, ms.GetCellID("B", nRow+encCount),
							ms.GetCellID("D", nRow+encCount), style)
						workbook.MergeCell(label, ms.GetCellID("B", nRow+encCount),
							ms.GetCellID("D", nRow+encCount))
						workbook.SetCellValue(label, ms.GetCellID("B", nRow+encCount),
							enclave)
					}
				}
				style = ms.Styles["platformSensorData"]
				workbook.SetCellStyle(label, ms.GetCellID("E", nRow+encCount),
					ms.GetCellID("E", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("E", nRow+encCount),
					ms.GetTimeString(premission))
				workbook.SetCellStyle(label, ms.GetCellID("F", nRow+encCount),
					ms.GetCellID("F", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("F", nRow+encCount),
					ms.GetTimeString(scheduled))
				workbook.SetCellStyle(label, ms.GetCellID("G", nRow+encCount),
					ms.GetCellID("G", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("G", nRow+encCount),
					ms.GetTimeString(executed))
				workbook.SetCellStyle(label, ms.GetCellID("H", nRow+encCount),
					ms.GetCellID("H", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("H", nRow+encCount),
					ms.GetTimeString(postmission))

				for _, outage := range outages {
					if strings.EqualFold(outage.GroundSystem, gs.ID) &&
						strings.EqualFold(outage.Classification, enclave) {
						outageNumber++
						outageTime += outage.OutageMinutes
					}
				}
				ao := 0.0
				if executed > 0 {
					ao = ((float64(executed) - float64(outageTime)) / float64(executed)) *
						100.0
				}
				style = ms.Styles["platformSensorData"]
				if ao > 0.0 && ao < 100.0 {
					style = ms.Styles["platformSensorDataNotNorm"]
				}
				workbook.SetCellStyle(label, ms.GetCellID("I", nRow+encCount),
					ms.GetCellID("I", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("I", nRow+encCount),
					strconv.FormatInt(int64(outageNumber), 10))
				workbook.SetCellStyle(label, ms.GetCellID("J", nRow+encCount),
					ms.GetCellID("J", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("J", nRow+encCount),
					ms.GetTimeString(outageTime))

				style = ms.Styles["platformSensorDataNum"]
				if ao > 0.0 && ao < 100.0 {
					style = ms.Styles["platformSensorDataNumNotNorm"]
				}

				workbook.SetCellStyle(label, ms.GetCellID("K", nRow+encCount),
					ms.GetCellID("K", nRow+encCount), style)
				workbook.SetCellValue(label, ms.GetCellID("K", nRow+encCount),
					ao)
			}
			nRow += encCount
		}
	}
	delta := end.Sub(start)
	if rptType == systemdata.ALL && delta > (24*time.Hour) {
		ms.AddSummarySheet(workbook, start, end, false, systemdata.GEOINT, "GEOINT")
	}
	if daily {
		dayStart := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
			time.UTC)
		for tm := dayStart; tm.Before(end); tm = tm.Add(24 * time.Hour) {
			dayEnd := tm.Add(24 * time.Hour).Add(-1 * time.Second)
			ms.AddSummarySheet(workbook, tm, dayEnd, false, rptType, "")
		}
	}
}

func (ms *MissionSummary) GetDateString(date time.Time) string {
	fmt.Println(date.Local())
	return date.Local().Format("01/02/2006")
}

func (ms *MissionSummary) GetTimeString(minutes uint) string {
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

func (ms *MissionSummary) GetCellID(letter string, row int) string {
	return letter + strconv.FormatInt(int64(row), 10)
}

func (ms *MissionSummary) GetTotalScheduled(exploit string,
	mtypes map[string]MissionType, sens []string) uint {
	answer := uint(0)
	for _, mtype := range mtypes {
		answer += uint(mtype.GetScheduled(exploit, sens))
	}
	return answer
}

func (ms *MissionSummary) GetTotalExecuted(exploit string,
	mtypes map[string]MissionType, sens []string) uint {
	answer := uint(0)
	for _, mtype := range mtypes {
		answer += uint(mtype.GetExecuted(exploit, sens))
	}
	return answer
}

func (ms *MissionSummary) GetTotalCancelled(exploit string,
	mtypes map[string]MissionType, sens []string) uint {
	answer := uint(0)
	for _, mtype := range mtypes {
		answer += uint(mtype.GetCancelled(exploit, sens))
	}
	return answer
}

func (ms *MissionSummary) GetTotalAborted(exploit string,
	mtypes map[string]MissionType, sens []string) uint {
	answer := uint(0)
	for _, mtype := range mtypes {
		answer += uint(mtype.GetAborted(exploit, sens))
	}
	return answer
}
