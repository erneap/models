package sites

import "github.com/erneap/go-models/employees"

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
