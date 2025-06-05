package sites

import (
	"sort"
	"strings"
	"time"

	"github.com/erneap/models/v2/employees"
)

type Shift struct {
	ID              string               `json:"id" bson:"id"`
	Name            string               `json:"name" bson:"name"`
	SortID          uint                 `json:"sort" bson:"sort"`
	AssociatedCodes []string             `json:"associatedCodes,omitempty" bson:"associatedCodes,omitempty"`
	PayCode         uint                 `json:"payCode" bson:"payCode"`
	Minimums        uint                 `json:"minimums" bson:"minimums"`
	Employees       []employees.Employee `json:"-" bson:"_"`
}

type ByShift []Shift

func (c ByShift) Len() int { return len(c) }
func (c ByShift) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByShift) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Position struct {
	ID        string               `json:"id" bson:"id"`
	Name      string               `json:"name" bson:"name"`
	SortID    uint                 `json:"sort" bson:"sort"`
	Assigned  []string             `json:"assigned" bson:"assigned"`
	Employees []employees.Employee `json:"-" bson:"_"`
}

type ByPosition []Position

func (c ByPosition) Len() int { return len(c) }
func (c ByPosition) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByPosition) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Workcenter struct {
	ID        string     `json:"id" bson:"id"`
	Name      string     `json:"name" bson:"name"`
	SortID    uint       `json:"sort" bson:"sort"`
	Shifts    []Shift    `json:"shifts,omitempty" bson:"shifts,omitempty"`
	Positions []Position `json:"positions,omitempty" bson:"positions,omitempty"`
}

type ByWorkcenter []Workcenter

func (c ByWorkcenter) Len() int { return len(c) }
func (c ByWorkcenter) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByWorkcenter) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (w *Workcenter) Assign(e *employees.Employee, date time.Time) {
	// check if employee is assigned to a position
	bPosition := false
	if len(w.Positions) > 0 {
		for p, pos := range w.Positions {
			for _, asgn := range pos.Assigned {
				if strings.EqualFold(asgn, e.ID.Hex()) {
					bPosition = true
					pos.Employees = append(pos.Employees, *e)
					sort.Sort(employees.ByEmployees(pos.Employees))
					w.Positions[p] = pos
				}
			}
		}
	}
	if !bPosition && len(w.Shifts) > 0 {
		wc := e.GetWorkday(date, date)
		for s, shft := range w.Shifts {
			for _, code := range shft.AssociatedCodes {
				if strings.EqualFold(wc.Code, code) {
					shft.Employees = append(shft.Employees, *e)
					w.Shifts[s] = shft
				}
			}
		}
	}
}

func (w *Workcenter) ClearEmployees() {
	for p, pos := range w.Positions {
		pos.Employees = pos.Employees[:0]
		w.Positions[p] = pos
	}
	for s, shft := range w.Shifts {
		shft.Employees = shft.Employees[:0]
		w.Shifts[s] = shft
	}
}
