package employees

import (
	"sort"
	"time"
)

type AnnualLeave struct {
	Year      int     `json:"year" bson:"year"`
	Annual    float64 `json:"annual" bson:"annual"`
	Carryover float64 `json:"carryover" bson:"carryover"`
}

type ByBalance []AnnualLeave

func (c ByBalance) Len() int { return len(c) }
func (c ByBalance) Less(i, j int) bool {
	return c[i].Year < c[j].Year
}
func (c ByBalance) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type LeaveDay struct {
	ID        int       `json:"id" bson:"id"`
	LeaveDate time.Time `json:"leavedate" bson:"leavedate"`
	Code      string    `json:"code" bson:"code"`
	Hours     float64   `json:"hours" bson:"hours"`
	Status    string    `json:"status" bson:"status"`
	RequestID string    `json:"requestid" bson:"requestid"`
	Used      bool      `json:"-" bson:"-"`
	TagDay    string    `json:"tagday,omitempty" bson:"tagday,omitempty"`
}

type ByLeaveDay []LeaveDay

func (c ByLeaveDay) Len() int { return len(c) }
func (c ByLeaveDay) Less(i, j int) bool {
	return c[i].LeaveDate.Before(c[j].LeaveDate)
}
func (c ByLeaveDay) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type LeaveRequestComment struct {
	CommentDate time.Time `json:"commentdate" bson:"commentdate"`
	Comment     string    `json:"comment" bson:"comment"`
}

type ByLeaveRequestComment []LeaveRequestComment

func (c ByLeaveRequestComment) Len() int { return len(c) }
func (c ByLeaveRequestComment) Less(i, j int) bool {
	return c[i].CommentDate.Before(c[j].CommentDate)
}
func (c ByLeaveRequestComment) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type LeaveRequest struct {
	ID            string                `json:"id" bson:"id"`
	EmployeeID    string                `json:"employeeid,omitempty"`
	RequestDate   time.Time             `json:"requestDate" bson:"requestDate"`
	PrimaryCode   string                `json:"primarycode" bson:"primarycode"`
	StartDate     time.Time             `json:"startdate" bson:"startdate"`
	EndDate       time.Time             `json:"enddate" bson:"enddate"`
	Status        string                `json:"status" bson:"status"`
	ApprovedBy    string                `json:"approvedby" bson:"approvedby"`
	ApprovalDate  time.Time             `json:"approvalDate" bson:"approvalDate"`
	RequestedDays []LeaveDay            `json:"requesteddays" bson:"requesteddays"`
	Comments      []LeaveRequestComment `json:"comments,omitempty" bson:"comments,omitempty"`
}

type ByLeaveRequest []LeaveRequest

func (c ByLeaveRequest) Len() int { return len(c) }
func (c ByLeaveRequest) Less(i, j int) bool {
	if c[i].StartDate.Equal(c[j].StartDate) {
		if c[i].EndDate.Equal(c[j].EndDate) {
			return c[i].ID < c[j].ID
		}
		return c[i].EndDate.Before(c[j].EndDate)
	}
	return c[i].StartDate.Before(c[j].StartDate)
}
func (c ByLeaveRequest) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (lr *LeaveRequest) SetLeaveDay(date time.Time, code string, hours float64) {
	for _, lv := range lr.RequestedDays {
		if lv.LeaveDate.Equal(date) {
			lv.Code = code
			lv.Hours = hours
		}
	}
}

func (lr *LeaveRequest) SetLeaveDays(emp *Employee) {
	sDate := time.Date(lr.StartDate.Year(), lr.StartDate.Month(),
		lr.StartDate.Day(), 0, 0, 0, 0, time.UTC)
	lr.RequestedDays = lr.RequestedDays[:0]
	for sDate.Before(lr.EndDate) || sDate.Equal(lr.EndDate) {
		wd := emp.GetWorkday(sDate, lr.StartDate.AddDate(0, 0, -1))
		if wd != nil && wd.Code != "" {
			hours := wd.Hours
			if lr.PrimaryCode == "H" {
				hours = 8.0
			}
			lv := LeaveDay{
				LeaveDate: sDate,
				Code:      lr.PrimaryCode,
				Hours:     hours,
				Status:    "REQUESTED",
				RequestID: lr.ID,
			}
			lr.RequestedDays = append(lr.RequestedDays, lv)
		}
		sDate = sDate.AddDate(0, 0, 1)
	}
	sort.Sort(ByLeaveDay(lr.RequestedDays))
}
