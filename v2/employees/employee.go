package employees

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/erneap/models/v2/labor"
	"github.com/erneap/models/v2/users"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Employee struct {
	ID             primitive.ObjectID  `json:"id" bson:"_id"`
	TeamID         primitive.ObjectID  `json:"team" bson:"team"`
	SiteID         string              `json:"site" bson:"site"`
	UserID         primitive.ObjectID  `json:"userid" bson:"userid"`
	Email          string              `json:"email" bson:"email"`
	Name           EmployeeName        `json:"name" bson:"name"`
	Data           *EmployeeData       `json:"data,omitempty" bson:"data,omitempty"`
	CompanyInfo    CompanyInfo         `json:"companyinfo"`
	Assignments    []Assignment        `json:"assignments,omitempty"`
	Variations     []Variation         `json:"variations,omitempty"`
	Balances       []AnnualLeave       `json:"balance,omitempty"`
	Leaves         []LeaveDay          `json:"leaves,omitempty"`
	Requests       []LeaveRequest      `json:"requests,omitempty"`
	LaborCodes     []EmployeeLaborCode `json:"laborCodes,omitempty"`
	User           *users.User         `json:"user,omitempty" bson:"-"`
	Work           []Work              `json:"work,omitempty" bson:"-"`
	ContactInfo    []Contact           `json:"contactinfo,omitempty" bson:"contactinfo,omitempty"`
	Specialties    []Specialty         `json:"specialties,omitempty" bson:"specialties,omitempty"`
	EmailAddresses []string            `json:"emails,omitempty" bson:"emails,omitempty"`
}

type ByEmployees []Employee

func (c ByEmployees) Len() int { return len(c) }
func (c ByEmployees) Less(i, j int) bool {
	if c[i].Name.LastName == c[j].Name.LastName {
		if c[i].Name.FirstName == c[j].Name.FirstName {
			return c[i].Name.MiddleName < c[j].Name.MiddleName
		}
		return c[i].Name.FirstName < c[j].Name.FirstName
	}
	return c[i].Name.LastName < c[j].Name.LastName
}
func (c ByEmployees) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type ByEmployeesFirst []Employee

func (c ByEmployeesFirst) Len() int { return len(c) }
func (c ByEmployeesFirst) Less(i, j int) bool {
	if c[i].Name.FirstName == c[j].Name.FirstName {
		if c[i].Name.LastName == c[j].Name.LastName {
			return c[i].Name.MiddleName < c[j].Name.MiddleName
		}
		return c[i].Name.LastName < c[j].Name.LastName
	}
	return c[i].Name.FirstName < c[j].Name.FirstName
}
func (c ByEmployeesFirst) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (e *Employee) RemoveLeaves(start, end time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	sort.Sort(ByLeaveDay(e.Leaves))
	startpos := -1
	endpos := -1
	for i, lv := range e.Leaves {
		if startpos < 0 && (lv.LeaveDate.Equal(start) || lv.LeaveDate.After(start)) &&
			(lv.LeaveDate.Equal(end) || lv.LeaveDate.Before(end)) {
			startpos = i
		} else if startpos >= 0 && (lv.LeaveDate.Equal(start) || lv.LeaveDate.After(start)) &&
			(lv.LeaveDate.Equal(end) || lv.LeaveDate.Before(end)) {
			endpos = i
		}
	}
	if startpos >= 0 {
		if endpos < 0 {
			endpos = startpos
		}
		e.Leaves = append(e.Leaves[:startpos], e.Leaves[endpos+1:]...)
	}
}

func (e *Employee) ConvertFromData() error {
	if e.Data != nil {
		e.CompanyInfo = e.Data.CompanyInfo
		e.Leaves = e.Data.Leaves
		e.Assignments = e.Data.Assignments
		e.Variations = e.Data.Variations
		e.Balances = e.Data.Balances
		e.Requests = e.Data.Requests
		for _, lc := range e.Data.LaborCodes {
			for a, asgmt := range e.Assignments {
				newLc := &EmployeeLaborCode{
					ChargeNumber: lc.ChargeNumber,
					Extension:    lc.Extension,
				}
				asgmt.LaborCodes = append(asgmt.LaborCodes, *newLc)
				e.Assignments[a] = asgmt
			}
		}
		e.Data = nil
	}
	return nil
}

type EmployeeName struct {
	FirstName  string `json:"first"`
	MiddleName string `json:"middle"`
	LastName   string `json:"last"`
	Suffix     string `json:"suffix"`
}

func (en *EmployeeName) GetLastFirst() string {
	return en.LastName + ", " + en.FirstName
}

func (en *EmployeeName) GetLastFirstMI() string {
	if en.MiddleName != "" {
		return en.LastName + ", " + en.FirstName + " " + en.MiddleName[0:1]
	}
	return en.LastName + ", " + en.FirstName
}

type EmployeeData struct {
	CompanyInfo CompanyInfo         `json:"companyinfo"`
	Assignments []Assignment        `json:"assignments,omitempty"`
	Variations  []Variation         `json:"variations,omitempty"`
	Balances    []AnnualLeave       `json:"balance,omitempty"`
	Leaves      []LeaveDay          `json:"leaves,omitempty"`
	Requests    []LeaveRequest      `json:"requests,omitempty"`
	LaborCodes  []EmployeeLaborCode `json:"laborCodes,omitempty"`
}

func (e *Employee) IsActive(date time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if asgmt.UseAssignment(e.SiteID, date) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) IsAssigned(site, workcenter string, start, end time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if strings.EqualFold(asgmt.Site, site) &&
			strings.EqualFold(asgmt.Workcenter, workcenter) &&
			asgmt.StartDate.After(end) && asgmt.EndDate.Before((start)) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) AtSite(site string, start, end time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := false
	for _, asgmt := range e.Assignments {
		if strings.EqualFold(asgmt.Site, site) &&
			asgmt.StartDate.Before(end) && asgmt.EndDate.After((start)) {
			answer = true
		}
	}
	return answer
}

func (e *Employee) GetWorkday(date, lastWork time.Time) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	work := 0.0
	stdWorkDay := 8.0
	for _, asgmt := range e.Assignments {
		if asgmt.UseAssignment(e.SiteID, date) {
			stdWorkDay = asgmt.GetStandardWorkday()
		}
	}
	var siteid string = ""
	for _, wk := range e.Work {
		if wk.DateWorked.Year() == date.Year() &&
			wk.DateWorked.Month() == date.Month() &&
			wk.DateWorked.Day() == date.Day() && !wk.ModifiedTime {
			work += wk.Hours
		}
	}
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date)
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	if work > 0.0 {
		for wkday == nil || wkday.Code == "" {
			date = date.AddDate(0, 0, -1)
			for _, asgmt := range e.Assignments {
				if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
					(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
					wkday = asgmt.GetWorkday(date)
				}
			}
		}
		return wkday
	}
	if date.Equal(lastWork) || date.Before(lastWork) {
		wkday = nil
	}

	for _, lv := range e.Leaves {
		if lv.LeaveDate.Year() == date.Year() &&
			lv.LeaveDate.Month() == date.Month() &&
			lv.LeaveDate.Day() == date.Day() &&
			(lv.Hours > (stdWorkDay/2) || (work == 0.0 &&
				strings.EqualFold(lv.Status, "actual"))) {
			wkday = &Workday{
				ID:         uint(0),
				Workcenter: "",
				Code:       lv.Code,
				Hours:      lv.Hours,
			}
		}
	}
	return wkday
}

func (e *Employee) GetWorkdayActual(date time.Time,
	labor []EmployeeLaborCode) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	var siteid string = ""
	bPrimary := false
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date)
			for _, lc := range labor {
				for _, alc := range asgmt.LaborCodes {
					if strings.EqualFold(lc.ChargeNumber, alc.ChargeNumber) &&
						strings.EqualFold(lc.Extension, alc.Extension) {
						bPrimary = true
					}
				}
			}
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	bLeave := false
	if bPrimary || len(labor) == 0 {
		for _, lv := range e.Leaves {
			if lv.LeaveDate.Equal(date) &&
				strings.EqualFold(lv.Status, "actual") {
				if !bLeave {
					wkday = &Workday{
						ID:         uint(0),
						Workcenter: "",
						Code:       lv.Code,
						Hours:      lv.Hours,
					}
					bLeave = true
				} else {
					if lv.Hours <= wkday.Hours {
						wkday.Hours += lv.Hours
					} else {
						wkday.Hours += lv.Hours
						wkday.Code = lv.Code
					}
				}
			}
		}
	}
	return wkday
}

func (e *Employee) GetWorkdayWOLeave(date time.Time) *Workday {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var wkday *Workday = nil
	var siteid string = ""
	for _, asgmt := range e.Assignments {
		if (asgmt.StartDate.Before(date) || asgmt.StartDate.Equal(date)) &&
			(asgmt.EndDate.After(date) || asgmt.EndDate.Equal(date)) {
			siteid = asgmt.Site
			wkday = asgmt.GetWorkday(date)
		}
	}
	for _, vari := range e.Variations {
		if (vari.StartDate.Before(date) || vari.StartDate.Equal(date)) &&
			(vari.EndDate.After(date) || vari.EndDate.Equal(date)) {
			wkday = vari.GetWorkday(siteid, date)
		}
	}
	return wkday
}

func (e *Employee) GetStandardWorkday(date time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	lastWork := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	answer := 8.0
	count := 0
	start := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0,
		time.UTC)
	end := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	for start.Weekday() != time.Sunday {
		start = start.AddDate(0, 0, -1)
	}
	for end.Weekday() != time.Saturday {
		end = end.AddDate(0, 0, 1)
	}
	for start.Before(end) || start.Equal(end) {
		wd := e.GetWorkday(start, lastWork)
		if wd != nil && wd.Code != "" {
			count++
		}
		start = start.AddDate(0, 0, 1)
	}
	if count < 5 {
		answer = 10.0
	}
	return answer
}

func (e *Employee) AddAssignment(site, wkctr string, start time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// get next assignment id as one plus the highest in employee data
	max := 0
	for _, asgmt := range e.Assignments {
		if int(asgmt.ID) > max {
			max = int(asgmt.ID)
		}
	}

	// set the current highest or last end date to one day before
	// this assignment start date
	sort.Sort(ByAssignment(e.Assignments))
	if len(e.Assignments) > 0 {
		lastAsgmt := e.Assignments[len(e.Assignments)-1]
		lastAsgmt.EndDate = start.AddDate(0, 0, -1)
		e.Assignments[len(e.Assignments)-1] = lastAsgmt
	}

	// create the new assignment
	newAsgmt := Assignment{
		ID:           uint(max + 1),
		Site:         site,
		Workcenter:   wkctr,
		StartDate:    start,
		EndDate:      time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC),
		RotationDate: time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC),
		RotationDays: 0,
	}
	// add a single schedule, plus it's seven workdays, set schedule
	// automatically to M-F/workcenter/8 hours/day shift.
	newAsgmt.AddSchedule(7)
	for i, wd := range newAsgmt.Schedules[0].Workdays {
		if i != 0 && i != 6 {
			wd.Code = "D"
			wd.Workcenter = wkctr
			wd.Hours = 8.0
			newAsgmt.Schedules[0].Workdays[i] = wd
		}
	}

	// add it employees assignment list and sort them
	e.Assignments = append(e.Assignments, newAsgmt)
	sort.Sort(ByAssignment(e.Assignments))
}

func (e *Employee) RemoveAssignment(id uint) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	pos := -1
	if id > 1 {
		sort.Sort(ByAssignment(e.Assignments))
		for i, asgmt := range e.Assignments {
			if asgmt.ID == id {
				pos = i
			}
		}
		if pos >= 0 {
			asgmt := e.Assignments[pos-1]
			asgmt.EndDate = time.Date(9999, 12, 30, 0, 0, 0, 0, time.UTC)
			e.Assignments[pos-1] = asgmt
			e.Assignments = append(e.Assignments[:pos],
				e.Assignments[pos+1:]...)
		}
	}
}

func (e *Employee) IsPrimaryCode(date time.Time, chgno, ext string) bool {
	answer := false
	for _, asgmt := range e.Assignments {
		if asgmt.UseAssignment(e.SiteID, date) {
			for _, lc := range asgmt.LaborCodes {
				if strings.EqualFold(chgno, lc.ChargeNumber) &&
					strings.EqualFold(ext, lc.Extension) {
					answer = true
				}
			}
		}
	}
	return answer
}

func (e *Employee) PurgeOldData(date time.Time) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// purge old variations based on variation end date
	sort.Sort(ByVariation(e.Variations))
	for i := len(e.Variations) - 1; i >= 0; i-- {
		if e.Variations[i].EndDate.Before(date) {
			e.Variations = append(e.Variations[:i],
				e.Variations[i+1:]...)
		}
	}

	// purge old leave and leave requests based on leave date and
	// leave request end date.
	sort.Sort(ByLeaveDay(e.Leaves))
	sort.Sort(ByLeaveRequest(e.Requests))
	for i := len(e.Leaves) - 1; i >= 0; i-- {
		if e.Leaves[i].LeaveDate.Before(date) {
			e.Leaves = append(e.Leaves[:i], e.Leaves[i+1:]...)
		}
	}
	for i := len(e.Requests) - 1; i >= 0; i-- {
		if e.Requests[i].EndDate.Before(date) {
			e.Requests = append(e.Requests[:i], e.Requests[i+1:]...)
		}
	}

	// purge old leave balances based on year
	sort.Sort(ByBalance(e.Balances))
	for i := len(e.Balances) - 1; i >= 0; i-- {
		if e.Balances[i].Year < date.Year() {
			e.Balances = append(e.Balances[:i], e.Balances[i+1:]...)
		}
	}

	// check if employee quit before purge date
	sort.Sort(ByAssignment(e.Assignments))
	asgmt := e.Assignments[len(e.Assignments)-1]
	return asgmt.EndDate.Before(date)
}

func (e *Employee) CreateLeaveBalance(year int) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	lastAnnual := 0.0
	lastCarry := 0.0
	for _, al := range e.Balances {
		if al.Year == year {
			found = true
		}
		if al.Year == year-1 {
			lastAnnual = al.Annual
			lastCarry = al.Carryover
		}
	}
	if !found {
		al := AnnualLeave{
			Year:      year,
			Annual:    lastAnnual,
			Carryover: 0.0,
		}
		if lastAnnual == 0.0 {
			al.Annual = 120.0
		} else {
			carry := lastAnnual + lastCarry
			for _, lv := range e.Leaves {
				if lv.LeaveDate.Year() == year-1 && strings.ToLower(lv.Code) == "v" &&
					strings.ToLower(lv.Status) == "actual" {
					carry -= lv.Hours
				}
			}
			al.Carryover = carry
		}
		e.Balances = append(e.Balances, al)
	}
}

func (e *Employee) UpdateAnnualLeave(year int, annual, carry float64) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	for _, al := range e.Balances {
		if al.Year == year {
			found = true
			al.Annual = annual
			al.Carryover = carry
		}
	}
	if !found {
		al := AnnualLeave{
			Year:      year,
			Annual:    annual,
			Carryover: carry,
		}
		e.Balances = append(e.Balances, al)
		sort.Sort(ByBalance(e.Balances))
	}
}

func (e *Employee) AddLeave(id int, date time.Time, code, status string,
	hours float64, requestID *primitive.ObjectID) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	max := 0
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.Equal(date) &&
			strings.EqualFold(lv.Code, code)) || lv.ID == id {
			found = true
			lv.Status = status
			lv.Hours = hours
			if requestID != nil {
				lv.RequestID = requestID.Hex()
			}
		} else if lv.ID > max {
			max = lv.ID
		}
	}
	if !found {
		lv := LeaveDay{
			ID:        max + 1,
			LeaveDate: date,
			Code:      code,
			Hours:     hours,
			Status:    status,
			RequestID: requestID.Hex(),
		}
		e.Leaves = append(e.Leaves, lv)
		sort.Sort(ByLeaveDay(e.Leaves))
	}
}

func (e *Employee) UpdateLeave(id int, field, value string) (*LeaveDay, error) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var oldLv *LeaveDay
	oldLv = nil
	found := false
	for i := 0; i < len(e.Leaves) && !found; i++ {
		lv := e.Leaves[i]
		if lv.ID == id {
			oldLv = &LeaveDay{
				ID: lv.ID,
				LeaveDate: time.Date(lv.LeaveDate.Year(), lv.LeaveDate.Month(),
					lv.LeaveDate.Day(), 0, 0, 0, 0, time.UTC),
				Code:      lv.Code,
				Hours:     lv.Hours,
				Status:    lv.Status,
				RequestID: lv.RequestID,
				TagDay:    lv.TagDay,
			}
			switch strings.ToLower(field) {
			case "date":
				date, err := time.ParseInLocation("01/02/2006", value, time.UTC)
				if err != nil {
					return nil, err
				}
				lv.LeaveDate = date
			case "code":
				lv.Code = value
			case "hours":
				hrs, err := strconv.ParseFloat(value, 64)
				if err != nil {
					return nil, err
				}
				lv.Hours = hrs
			case "status":
				lv.Status = value
			case "requestid":
				lv.RequestID = value
			case "tagday":
				lv.TagDay = value
			}
			e.Leaves[i] = lv
		}
	}
	return oldLv, nil
}

func (e *Employee) DeleteLeave(id int) *LeaveDay {
	if e.Data != nil {
		e.ConvertFromData()
	}
	var oldLv *LeaveDay
	oldLv = nil
	pos := -1
	for i := 0; i < len(e.Leaves) && oldLv == nil; i++ {
		lv := e.Leaves[i]
		if lv.ID == id {
			oldLv = &lv
			pos = i
		}
	}
	if pos >= 0 {
		e.Leaves = append(e.Leaves[:pos], e.Leaves[pos+1:]...)
	}
	return oldLv
}

func (e *Employee) GetLeaveHours(start, end time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	sort.Sort(ByLeaveDay(e.Leaves))
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.After(start) ||
			lv.LeaveDate.Equal(start)) &&
			lv.LeaveDate.Before(end) &&
			strings.EqualFold(lv.Status, "actual") {
			answer += lv.Hours
		}
	}
	return answer
}

func (e *Employee) GetPTOHours(start, end time.Time) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	sort.Sort(ByLeaveDay(e.Leaves))
	for _, lv := range e.Leaves {
		if (lv.LeaveDate.After(start) ||
			lv.LeaveDate.Equal(start)) &&
			lv.LeaveDate.Before(end) &&
			strings.EqualFold(lv.Status, "actual") &&
			strings.EqualFold(lv.Code, "v") {
			answer += lv.Hours
		}
	}
	return answer
}

func (e *Employee) NewLeaveRequest(empID, code string, start, end time.Time,
	offset float64, comment string) *LeaveRequest {
	if e.Data != nil {
		e.ConvertFromData()
	}
	for l, lr := range e.Requests {
		if lr.StartDate.Equal(start) && lr.EndDate.Equal(end) {
			if comment != "" {
				lrc := &LeaveRequestComment{
					CommentDate: time.Now().UTC(),
					Comment:     comment,
				}
				lr.Comments = append(lr.Comments, *lrc)
				e.Requests[l] = lr
			}
			return &lr
		}
	}
	answer := &LeaveRequest{
		ID:          primitive.NewObjectID().Hex(),
		EmployeeID:  empID,
		RequestDate: time.Now().UTC(),
		PrimaryCode: code,
		StartDate:   start,
		EndDate:     end,
		Status:      "DRAFT",
	}
	if comment != "" {
		lrc := &LeaveRequestComment{
			CommentDate: time.Now().UTC(),
			Comment:     comment,
		}
		answer.Comments = append(answer.Comments, *lrc)
	}
	zoneID := "UTC"
	if offset > 0 {
		zoneID += "+" + fmt.Sprintf("%0.1f", offset)
	} else if offset < 0 {
		zoneID += fmt.Sprintf("%0.1f", offset)
	}
	sDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
		time.UTC)
	std := e.GetStandardWorkday(sDate)
	for !sDate.After(end) {
		wd := e.GetWorkday(sDate, start.AddDate(0, 0, -1))
		if wd != nil && wd.Code != "" {
			hours := wd.Hours
			if hours == 0.0 {
				hours = std
			}
			if code == "H" {
				hours = 8.0
			}
			lv := LeaveDay{
				LeaveDate: sDate,
				Code:      code,
				Hours:     hours,
				Status:    "DRAFT",
				RequestID: answer.ID,
			}
			answer.RequestedDays = append(answer.RequestedDays, lv)
		}
		sDate = sDate.AddDate(0, 0, 1)
	}
	e.Requests = append(e.Requests, *answer)
	sort.Sort(ByLeaveRequest(e.Requests))
	return answer
}

func (e *Employee) UpdateLeaveRequest(request, field, value string,
	offset float64) (string, *LeaveRequest, error) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	message := ""
	for i, req := range e.Requests {
		if req.ID == request {
			switch strings.ToLower(field) {
			case "startdate", "start":
				lvDate, err := time.Parse("2006-01-02", value)
				if err != nil {
					return "", nil, err
				}
				if lvDate.Before(req.StartDate) || lvDate.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: Starting date changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				} else if strings.EqualFold(req.Status, "approved") {
					message = fmt.Sprintf("Leave Request from %s: Starting date changed "+
						"to %s", e.Name.GetLastFirst(), lvDate.Format("2006-01-03"))
				}
				req.StartDate = lvDate
				if req.StartDate.After(req.EndDate) {
					req.EndDate = lvDate
				}
				// reset the leave dates
				req = e.resetLeaveDays(req.PrimaryCode, req)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "enddate", "end":
				lvDate, err := time.Parse("2006-01-02", value)
				if err != nil {
					return "", nil, err
				}
				if lvDate.Before(req.StartDate) || lvDate.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: Ending Date changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				} else if strings.EqualFold(req.Status, "approved") {
					message = fmt.Sprintf("Leave Request from %s: Ending Date changed "+
						"to %s", e.Name.GetLastFirst(), lvDate.Format("2006-01-02"))
				}
				req.EndDate = lvDate
				if req.EndDate.Before(req.StartDate) {
					req.StartDate = lvDate
				}
				// reset the leave dates
				req = e.resetLeaveDays(req.PrimaryCode, req)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "code", "primarycode":
				req.PrimaryCode = value
				if strings.EqualFold(value, req.PrimaryCode) {
					req = e.resetLeaveDays(value, req)
				}
			case "dates":
				parts := strings.Split(value, "|")
				start, err := time.ParseInLocation("2006-01-02", parts[0], time.UTC)
				if err != nil {
					return "", nil, err
				}
				end, err := time.ParseInLocation("2006-01-02", parts[1], time.UTC)
				if err != nil {
					return "", nil, err
				}
				start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
					time.UTC)
				end = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0,
					time.UTC)
				if start.Before(req.StartDate) || start.After(req.EndDate) ||
					end.Before(req.StartDate) || end.After(req.EndDate) {
					if strings.EqualFold(req.Status, "approved") {
						req.Status = "REQUESTED"
						req.ApprovalDate = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
						req.ApprovedBy = ""
						message = fmt.Sprintf("Leave Request from %s: dates changed "+
							"needs reapproval", e.Name.GetLastFirst())
					}
					startPos := -1
					endPos := -1
					sort.Sort(ByLeaveDay(e.Leaves))
					for i, lv := range e.Leaves {
						if lv.RequestID == req.ID {
							if startPos < 0 {
								startPos = i
							} else {
								endPos = i
							}
						}
					}
					if startPos >= 0 {
						if endPos < 0 {
							endPos = startPos
						}
						endPos++
						if endPos > len(e.Leaves) {

						} else {
							e.Leaves = append(e.Leaves[:startPos],
								e.Leaves[endPos:]...)
						}
					}
				}
				req.StartDate = time.Date(start.Year(), start.Month(), start.Day(), 0,
					0, 0, 0, time.UTC)
				req.EndDate = time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0,
					time.UTC)
				req = e.resetLeaveDays(req.PrimaryCode, req)
				if req.Status == "APPROVED" {
					e.ChangeApprovedLeaveDates(req)
				}
			case "requested":
				req.Status = "REQUESTED"
				for d, day := range req.RequestedDays {
					if day.Code != "" && day.Status == "" {
						day.Status = "REQUESTED"
					}
					req.RequestedDays[d] = day
				}
				message = fmt.Sprintf("Leave Request: Leave Request from %s ",
					e.Name.GetLastFirst()) + "submitted for approval.  " +
					fmt.Sprintf("Requested Leave Date: %s - %s.",
						req.StartDate.Format("02 Jan 06"), req.EndDate.Format("02 Jan 06"))
			case "unapprove":
				req.ApprovedBy = ""
				req.ApprovalDate = time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
				req.Status = "DRAFT"
				for d, day := range req.RequestedDays {
					day.Status = "REQUESTED"
					req.RequestedDays[d] = day
				}
				cmt := LeaveRequestComment{
					CommentDate: time.Now().UTC(),
					Comment:     value,
				}
				req.Comments = append(req.Comments, cmt)
				message = "Leave Request: Leave Request unapproved.\n" +
					"Comment: " + value
			case "day", "requestday":
				bApproved := strings.EqualFold(req.Status, "approved")
				parts := strings.Split(value, "|")
				lvDate, _ := time.Parse("2006-01-02", parts[0])
				code := parts[1]
				hours, _ := strconv.ParseFloat(parts[2], 64)
				found := false
				status := ""
				workcenter := ""
				if len(parts) > 3 {
					workcenter = parts[3]
				}
				for j, lv := range req.RequestedDays {
					if lv.LeaveDate.Equal(lvDate) {
						found = true
						lv.Code = code
						if status == "" {
							status = lv.Status
						}
						lv.Status = workcenter
						if code == "" {
							lv.Hours = 0.0
						} else {
							lv.Hours = hours
						}
						req.RequestedDays[j] = lv
					}
				}

				if !found {
					lv := LeaveDay{
						LeaveDate: lvDate,
						Code:      code,
						Hours:     hours,
						Status:    status,
						RequestID: req.ID,
					}
					req.RequestedDays = append(req.RequestedDays, lv)
				}
				if bApproved {
					found = false
					for j, lv := range e.Leaves {
						if lv.LeaveDate.Equal(lvDate) {
							found = true
							lv.Code = code
							if code == "" {
								lv.Hours = 0.0
							} else {
								lv.Hours = hours
							}
							e.Leaves[j] = lv
						}
					}
					if !found && code != "" {
						lv := &LeaveDay{
							LeaveDate: lvDate,
							Code:      code,
							Hours:     hours,
							Status:    req.Status,
							RequestID: req.ID,
						}
						e.Leaves = append(e.Leaves, *lv)
					}
				}
			case "comment", "addcomment":
				newComment := &LeaveRequestComment{
					CommentDate: time.Now().UTC(),
					Comment:     value,
				}
				req.Comments = append(req.Comments, *newComment)
			}
			e.Requests[i] = req
			return message, &req, nil
		}
	}
	return "", nil, errors.New("not found")
}

func (e *Employee) resetLeaveDays(value string, req LeaveRequest) LeaveRequest {
	if strings.ToLower(value) == "mod" {
		req.RequestedDays = req.RequestedDays[:0]
		start := time.Date(req.StartDate.Year(), req.StartDate.Month(),
			req.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		for start.Weekday() != time.Sunday {
			start = start.AddDate(0, 0, -1)
		}
		end := time.Date(req.EndDate.Year(), req.EndDate.Month(),
			req.EndDate.Day(), 0, 0, 0, 0, time.UTC)
		for end.Weekday() != time.Saturday {
			end = end.AddDate(0, 0, 1)
		}
		lastDay := e.GetLastWorkday()
		count := -1
		for start.Before(end) || start.Equal(end) {
			count++
			wd := e.GetWorkday(start, lastDay)
			day := LeaveDay{
				ID:        count,
				LeaveDate: start,
				Code:      wd.Code,
				Hours:     wd.Hours,
				Status:    wd.Workcenter,
			}
			req.RequestedDays = append(req.RequestedDays, day)
			start = start.AddDate(0, 0, 1)
		}
	} else {
		req.RequestedDays = req.RequestedDays[:0]
		start := time.Date(req.StartDate.Year(), req.StartDate.Month(),
			req.StartDate.Day(), 0, 0, 0, 0, time.UTC)
		end := time.Date(req.EndDate.Year(), req.EndDate.Month(),
			req.EndDate.Day(), 0, 0, 0, 0, time.UTC)
		lastDay := e.GetLastWorkday()
		count := -1
		hours := e.GetStandardWorkday(start)
		if strings.EqualFold(value, "h") {
			hours = 8.0
		}
		for start.Before(end) || start.Equal(end) {
			count++
			wd := e.GetWorkday(start, lastDay)
			if wd.Code != "" {
				day := LeaveDay{
					ID:        count,
					LeaveDate: start,
					Code:      value,
					Hours:     hours,
					Status:    "REQUESTED",
				}
				req.RequestedDays = append(req.RequestedDays, day)
			} else {
				day := LeaveDay{
					ID:        count,
					LeaveDate: start,
					Code:      "",
					Hours:     0.0,
					Status:    "REQUESTED",
				}
				req.RequestedDays = append(req.RequestedDays, day)
			}
			start = start.AddDate(0, 0, 1)
		}
	}
	return req
}

func (e *Employee) ApproveLeaveRequest(request, field, value string,
	offset float64, leavecodes []labor.Workcode) (string, *LeaveRequest, error) {

	if e.Data != nil {
		e.ConvertFromData()
	}
	message := ""
	for i, req := range e.Requests {
		if req.ID == request {
			req.ApprovedBy = value
			req.ApprovalDate = time.Now().UTC()
			req.Status = "APPROVED"
			maxLvID := 0
			// remove any leaves associated with this request
			var deletes []int
			for l, lv := range e.Leaves {
				if strings.ToLower(lv.Status) != "actual" && (lv.RequestID == req.ID ||
					((lv.LeaveDate.Equal(req.StartDate) || lv.LeaveDate.Equal(req.EndDate)) ||
						(lv.LeaveDate.After(req.StartDate) && lv.LeaveDate.Before(req.EndDate)))) {
					deletes = append(deletes, l)
				}
				if lv.ID > maxLvID {
					maxLvID = lv.ID
				}
			}
			if len(deletes) > 0 {
				for d := len(deletes) - 1; d >= 0; d-- {
					pos := deletes[d]
					e.Leaves = append(e.Leaves[:pos], e.Leaves[pos+1:]...)
				}
			}
			if strings.ToLower(req.PrimaryCode) != "mod" {
				for d, day := range req.RequestedDays {
					day.Status = "APPROVED"
					req.RequestedDays[d] = day
				}
				message = "Leave Request: Leave Request approved."
				e.ChangeApprovedLeaveDates(req)
			} else if strings.ToLower(req.PrimaryCode) == "mod" {
				// check for variation for period
				// if yes, modify variation
				found := false
				for v, vari := range e.Variations {
					if vari.StartDate.Equal(req.StartDate) &&
						vari.EndDate.Equal(req.EndDate) && vari.IsMod {
						found = true
						extra := int(req.StartDate.Weekday())
						lastCode := ""
						workcenter := ""
						for _, day := range req.RequestedDays {
							isLeave := false
							for _, wc := range leavecodes {
								if strings.EqualFold(wc.Id, day.Code) &&
									wc.IsLeave {
									isLeave = true
								}
							}
							if isLeave {
								maxLvID++
								lv := LeaveDay{
									ID:        maxLvID,
									LeaveDate: day.LeaveDate,
									Code:      day.Code,
									Hours:     day.Hours,
									Status:    "APPROVED",
									RequestID: req.ID,
								}
								e.Leaves = append(e.Leaves, lv)
								sort.Sort(ByLeaveDay(e.Leaves))
							} else {
								lastCode = day.Code
								workcenter = day.Status
							}
							dow := (int(day.LeaveDate.Weekday()) + extra)
							if dow < len(vari.Schedule.Workdays) {
								tday := vari.Schedule.Workdays[dow]
								tday.Code = lastCode
								tday.Hours = day.Hours
								tday.Workcenter = workcenter
								vari.Schedule.Workdays[dow] = tday
							} else {
								tday := Workday{
									ID:         uint(dow),
									Code:       lastCode,
									Workcenter: workcenter,
									Hours:      day.Hours,
								}
								vari.Schedule.Workdays = append(vari.Schedule.Workdays, tday)
							}
						}
						e.Variations[v] = vari
					}
				}
				// if no, create new variation
				if !found {
					site := e.SiteID
					max := uint(0)
					for _, vari := range e.Variations {
						if vari.ID > max {
							max = vari.ID
						}
					}
					vari := Variation{
						ID:        max + 1,
						IsMids:    false,
						IsMod:     true,
						StartDate: req.StartDate,
						EndDate:   req.EndDate,
						Site:      site,
					}
					vari.Schedule = Schedule{
						ID:        0,
						ShowDates: true,
					}
					vari.Schedule.Workdays = vari.Schedule.Workdays[:0]
					start := time.Date(req.StartDate.Year(), req.StartDate.Month(),
						req.StartDate.Day(), 0, 0, 0, 0, time.UTC)
					for start.Weekday() != time.Sunday {
						start = start.AddDate(0, 0, -1)
					}
					end := time.Date(req.EndDate.Year(), req.EndDate.Month(),
						req.EndDate.Day(), 0, 0, 0, 0, time.UTC)
					for end.Weekday() != time.Saturday {
						end = end.AddDate(0, 0, 1)
					}
					count := -1
					for start.Before(end) || start.Equal(end) {
						count++
						var day Workday
						day.ID = uint(count)
						found = false
						for _, d := range req.RequestedDays {
							if !found && d.LeaveDate.Year() == start.Year() &&
								d.LeaveDate.Month() == start.Month() &&
								d.LeaveDate.Day() == start.Day() {
								found = true
								day.Code = d.Code
								day.Hours = d.Hours
								day.Workcenter = d.Status
							}
						}
						vari.Schedule.Workdays = append(vari.Schedule.Workdays, day)
						start = start.AddDate(0, 0, 1)
					}
					e.Variations = append(e.Variations, vari)
					sort.Sort(ByVariation(e.Variations))
				}
			}
			e.Requests[i] = req
			return message, &req, nil
		}
	}
	return "", nil, errors.New("not found")
}

func (e *Employee) ChangeApprovedLeaveDates(lr LeaveRequest) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	// approved leave affects the leave listing, so we will
	// remove old leaves for the period then add the new ones
	startPos := -1
	endPos := -1
	maxId := -1
	sort.Sort(ByLeaveDay(e.Leaves))
	for i, lv := range e.Leaves {
		if (lv.LeaveDate.After(lr.StartDate) || lv.LeaveDate.Equal(lr.StartDate)) &&
			(lv.LeaveDate.Before(lr.EndDate) || lv.LeaveDate.Equal(lr.EndDate)) {
			if startPos < 0 {
				startPos = i
			} else {
				endPos = i
			}
		}
		if maxId < lv.ID {
			maxId = lv.ID
		}
	}
	if startPos > 0 {
		if endPos < 0 {
			endPos = startPos
		}
		endPos++
		e.Leaves = append(e.Leaves[:startPos], e.Leaves[endPos:]...)
	}

	// now add the leave request's leave days to the leave list. if now mod time
	for _, lv := range lr.RequestedDays {
		if lv.Hours > 0.0 {
			maxId++
			lv.ID = maxId
			lv.Status = lr.Status
			lv.RequestID = lr.ID
			e.Leaves = append(e.Leaves, lv)
		}
	}
	sort.Sort(ByLeaveDay(e.Leaves))
}

func (e *Employee) DeleteLeaveRequest(request string) (string, error) {
	message := ""
	if e.Data != nil {
		e.ConvertFromData()
	}
	pos := -1
	var deletable *LeaveRequest
	for i, req := range e.Requests {
		if req.ID == request {
			pos = i
			deletable = &req
			message = fmt.Sprintf("Deleted Leave Request for %s, Dates: %s to %s ",
				e.Name.GetLastFirst(), req.StartDate.Format("01/02/06"),
				req.EndDate.Format("01/02/06"))
		}
	}
	if pos < 0 {
		return "", errors.New("request not found")
	}
	e.Requests = append(e.Requests[:pos], e.Requests[pos+1:]...)
	// delete all leaves associated with this leave request, except if the leave
	// has a status of actual
	if strings.ToLower(deletable.PrimaryCode) != "mod" {
		sort.Sort(ByLeaveDay(e.Leaves))
		var deletes []int
		for i, lv := range e.Leaves {
			if lv.RequestID == request && strings.ToLower(lv.Status) != "actual" {
				deletes = append(deletes, i)
			}
		}
		if len(deletes) > 0 {
			for i := len(deletes) - 1; i >= 0; i-- {
				e.Leaves = append(e.Leaves[:deletes[i]],
					e.Leaves[deletes[i]+1:]...)
			}
		}
	} else {
		pos = -1
		for v, vari := range e.Variations {
			if vari.IsMod && vari.StartDate.Equal(deletable.StartDate) &&
				vari.EndDate.Equal(deletable.EndDate) {
				pos = v
			}
		}
		if pos >= 0 {
			e.Variations = append(e.Variations[:pos], e.Variations[pos+1:]...)
		}
	}
	return message, nil
}

func (e *Employee) HasLaborCode(chargeNumber, extension string) bool {
	if e.Data != nil {
		e.ConvertFromData()
	}
	found := false
	for _, asgmt := range e.Assignments {
		for _, lc := range asgmt.LaborCodes {
			if strings.EqualFold(lc.ChargeNumber, chargeNumber) &&
				strings.EqualFold(lc.Extension, extension) {
				found = true
			}
		}
	}
	return found
}

func (e *Employee) DeleteLaborCode(chargeNo, ext string) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	if e.HasLaborCode(chargeNo, ext) {
		for a, asgmt := range e.Assignments {
			pos := -1
			for i, lc := range asgmt.LaborCodes {
				if lc.ChargeNumber == chargeNo && lc.Extension == ext {
					pos = i
				}
			}
			if pos >= 0 {
				asgmt.LaborCodes = append(asgmt.LaborCodes[:pos], asgmt.LaborCodes[pos+1:]...)
				e.Assignments[a] = asgmt
			}
		}
	}
}

func (e *Employee) DeleteLeavesBetweenDates(start, end time.Time) {
	if e.Data != nil {
		e.ConvertFromData()
	}
	for i := len(e.Leaves) - 1; i >= 0; i-- {
		if e.Leaves[i].LeaveDate.Equal(start) ||
			e.Leaves[i].LeaveDate.Equal(end) ||
			(e.Leaves[i].LeaveDate.After(start) &&
				e.Leaves[i].LeaveDate.Before(end)) {
			e.Leaves = append(e.Leaves[:i], e.Leaves[i+1:]...)
		}
	}
}

func (e *Employee) GetAssignment(start, end time.Time) (string, string) {
	assigned := make(map[string]int)
	current := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0,
		time.UTC)
	for current.Before(end) {
		wd := e.GetWorkdayWOLeave(current)
		if wd != nil {
			label := wd.Workcenter + "-" + wd.Code
			if label != "-" {
				val, ok := assigned[label]
				if ok {
					assigned[label] = val + 1
				} else {
					assigned[label] = 1
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}
	max := 0
	answer := ""
	for k, v := range assigned {
		if v > max {
			answer = k
			max = v
		}
	}
	if answer != "" {
		parts := strings.Split(answer, "-")
		return parts[0], parts[1]
	}
	return "", ""
}

func (e *Employee) GetWorkedHours(start, end time.Time) float64 {
	answer := 0.0

	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) ||
			wk.DateWorked.After(start)) &&
			wk.DateWorked.Before(end) &&
			!wk.ModifiedTime {
			answer += wk.Hours
		}
	}

	return answer
}

func (e *Employee) GetWorkedHoursForLabor(chgno, ext string,
	start, end time.Time) float64 {
	answer := 0.0

	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) ||
			wk.DateWorked.After(start)) &&
			wk.DateWorked.Before(end) &&
			strings.EqualFold(chgno, wk.ChargeNumber) &&
			strings.EqualFold(ext, wk.Extension) {
			answer += wk.Hours
		}
	}
	return answer
}

func (e *Employee) GetForecastHours(lCode labor.LaborCode,
	start, end time.Time, workcodes []EmployeeCompareCode,
	offset float64) float64 {
	if e.Data != nil {
		e.ConvertFromData()
	}
	answer := 0.0

	// first check to see if assigned this labor code, if not
	// return 0 hours
	found := false
	for _, asgmt := range e.Assignments {
		for _, lc := range asgmt.LaborCodes {
			if strings.EqualFold(lCode.ChargeNumber, lc.ChargeNumber) &&
				strings.EqualFold(lCode.Extension, lc.Extension) {
				found = true
			}
		}
	}
	if !found {
		return 0.0
	}

	// determine if provided labor code is applicable in
	// period.
	if lCode.EndDate.Before(start) || lCode.StartDate.After(end) {
		return 0.0
	}

	// determine last day of actual recorded work so than
	// forecast hours don't overlap.
	lastWork := time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)
	if len(e.Work) > 0 {
		sort.Sort(ByEmployeeWork(e.Work))
		lastWork = e.Work[len(e.Work)-1].DateWorked
	}
	// check leave hours for actual being later than lastWork
	if len(e.Leaves) > 0 {
		sort.Sort(ByLeaveDay(e.Leaves))
		for _, lv := range e.Leaves {
			if strings.EqualFold(lv.Status, "actual") &&
				lv.LeaveDate.After(lastWork) {
				lastWork = lv.LeaveDate
			}
		}
	}

	// now step through the days of the period to:
	// 1) see if they had worked any charge numbers during
	//		the period, if working add 0 hours
	// 2) see if they were supposed to be working on this
	//		date, compare workday code to workcodes to ensure
	//		they weren't on leave.  If not on leave, add
	// 		standard work day.
	current := time.Date(start.Year(), start.Month(),
		start.Day(), 0, 0, 0, 0, time.UTC)
	for current.Before(end) {
		if current.After(lastWork) {
			hours := e.GetWorkedHours(current, current.AddDate(0, 0, 1))
			if hours == 0.0 {
				if current.Equal(lCode.StartDate) || current.Equal(lCode.EndDate) ||
					(current.After(lCode.StartDate) && current.Before(lCode.EndDate)) {
					wd := e.GetWorkday(current, lastWork)
					if wd != nil && wd.Code != "" {
						for _, wc := range workcodes {
							if strings.EqualFold(wc.Code, wd.Code) && !wc.IsLeave {
								std := e.GetStandardWorkday(current)
								for _, asgmt := range e.Assignments {
									if current.Equal(asgmt.StartDate) || current.Equal(asgmt.EndDate) ||
										(current.After(asgmt.StartDate) && current.Before(asgmt.EndDate)) {
										for _, lc := range asgmt.LaborCodes {
											if strings.EqualFold(lCode.ChargeNumber, lc.ChargeNumber) &&
												strings.EqualFold(lCode.Extension, lc.Extension) {
												answer += std
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
		current = current.AddDate(0, 0, 1)
	}

	return answer
}

func (e *Employee) GetLastWorkday() time.Time {
	if e.Data != nil {
		e.ConvertFromData()
	}
	sort.Sort(ByEmployeeWork(e.Work))
	answer := time.Date(1970, 1, 1, 0, 0, 0, 0, time.UTC)
	if len(e.Work) > 0 {
		work := e.Work[len(e.Work)-1]
		answer = time.Date(work.DateWorked.Year(), work.DateWorked.Month(),
			work.DateWorked.Day(), 0, 0, 0, 0, time.UTC)
	}
	return answer
}

func (e *Employee) AddContactInfo(typeID int, value string, sortid int) {
	found := false
	next := -1
	for c, contact := range e.ContactInfo {
		if next < contact.Id {
			next = contact.Id
		}
		if contact.TypeID == typeID {
			found = true
			contact.Value = value
			e.ContactInfo[c] = contact
		}
	}
	if !found {
		contact := &Contact{
			Id:     next + 1,
			TypeID: typeID,
			Value:  value,
			SortID: sortid,
		}
		e.ContactInfo = append(e.ContactInfo, *contact)
		sort.Sort(ByEmployeeContact(e.ContactInfo))
	}
}

func (e *Employee) ResortContactInfo(teamContacts map[int]int) {
	for c, contact := range e.ContactInfo {
		if val, ok := teamContacts[contact.TypeID]; ok {
			contact.SortID = val
		}
		e.ContactInfo[c] = contact
	}
	sort.Sort(ByEmployeeContact(e.ContactInfo))
}

func (e *Employee) DeleteContactInfoByType(id int) {
	pos := -1
	for c, contact := range e.ContactInfo {
		if contact.TypeID == id {
			pos = c
		}
	}
	if pos >= 0 {
		e.ContactInfo = append(e.ContactInfo[:pos], e.ContactInfo[pos+1:]...)
	}
	sort.Sort(ByEmployeeContact(e.ContactInfo))
}

func (e *Employee) DeleteContactInfo(id int) {
	pos := -1
	for c, contact := range e.ContactInfo {
		if contact.Id == id {
			pos = c
		}
	}
	if pos >= 0 {
		e.ContactInfo = append(e.ContactInfo[:pos], e.ContactInfo[pos+1:]...)
	}
	sort.Sort(ByEmployeeContact(e.ContactInfo))
}

func (e *Employee) AddSpecialty(specID int, qualified bool, sortid int) {
	found := false
	next := -1
	for s, specialty := range e.Specialties {
		if next < specialty.Id {
			next = specialty.Id
		}
		if specialty.SpecialtyID == specID {
			found = true
			specialty.Qualified = qualified
			e.Specialties[s] = specialty
		}
	}
	if !found {
		specialty := &Specialty{
			Id:          next + 1,
			SpecialtyID: specID,
			Qualified:   qualified,
			SortID:      sortid,
		}
		e.Specialties = append(e.Specialties, *specialty)
	}
	sort.Sort(ByEmployeeSpecialty(e.Specialties))
}

func (e *Employee) ResortSpecialties(specialties map[int]int) {
	for s, spec := range e.Specialties {
		if val, ok := specialties[spec.SpecialtyID]; ok {
			spec.SortID = val
		}
		e.Specialties[s] = spec
	}
	sort.Sort(ByEmployeeSpecialty(e.Specialties))
}

func (e *Employee) DeleteSpecialty(id int) {
	pos := -1
	for s, spec := range e.Specialties {
		if spec.Id == id {
			pos = s
		}
	}
	if pos >= 0 {
		e.Specialties = append(e.Specialties[:pos], e.Specialties[pos+1:]...)
	}
	sort.Sort(ByEmployeeSpecialty(e.Specialties))
}

func (e *Employee) DeleteSpecialtyByType(id int) {
	pos := -1
	for s, spec := range e.Specialties {
		if spec.SpecialtyID == id {
			pos = s
		}
	}
	if pos >= 0 {
		e.Specialties = append(e.Specialties[:pos], e.Specialties[pos+1:]...)
	}
	sort.Sort(ByEmployeeSpecialty(e.Specialties))
}

func (e *Employee) HasSpecialty(spec int) bool {
	answer := false
	for _, sp := range e.Specialties {
		if sp.SpecialtyID == spec {
			answer = true
		}
	}
	return answer
}

func (e *Employee) AddEmailAddress(email string) {
	found := false
	for _, em := range e.EmailAddresses {
		if strings.EqualFold(em, email) {
			found = true
		}
	}
	if !found {
		e.EmailAddresses = append(e.EmailAddresses, email)
		sort.Strings(e.EmailAddresses)
	}
}

func (e *Employee) RemoveEmailAddress(email string) {
	found := -1
	for e, em := range e.EmailAddresses {
		if strings.EqualFold(em, email) {
			found = e
		}
	}
	if found >= 0 {
		e.EmailAddresses = append(e.EmailAddresses[:found], e.EmailAddresses[found+1:]...)
	}
}

func (e *Employee) HasModTime(start, end time.Time) bool {
	answer := false
	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) || wk.DateWorked.Equal(end) ||
			(wk.DateWorked.After(start) && wk.DateWorked.Before(end))) &&
			wk.ModifiedTime {
			answer = true
		}
	}
	return answer
}

func (e *Employee) GetModTime(start, end time.Time) float64 {
	answer := 0.0
	for _, wk := range e.Work {
		if (wk.DateWorked.Equal(start) || wk.DateWorked.Equal(end) ||
			(wk.DateWorked.After(start) && wk.DateWorked.Before(end))) &&
			wk.ModifiedTime {
			answer += wk.Hours
		}
	}
	return answer
}

type EmployeeCompareCode struct {
	Code    string
	IsLeave bool
}
