package teams

import (
	"sort"
	"strings"
	"time"

	"github.com/erneap/go-models/labor"
	"github.com/erneap/go-models/sites"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Team struct {
	ID             primitive.ObjectID `json:"id" bson:"_id"`
	Name           string             `json:"name" bson:"name"`
	Workcodes      []labor.Workcode   `json:"workcodes" bson:"workcodes"`
	Sites          []sites.Site       `json:"sites" bson:"sites"`
	Companies      []Company          `json:"companies,omitempty" bson:"companies,omitempty"`
	ContactTypes   []ContactType      `json:"contacttypes,omitempty" bson:"contacttypes,omitempty"`
	SpecialtyTypes []SpecialtyType    `json:"specialties,omitempty" bson:"specialties,omitempty"`
}

type ByTeam []Team

func (c ByTeam) Len() int { return len(c) }
func (c ByTeam) Less(i, j int) bool {
	return c[i].Name < c[j].Name
}
func (c ByTeam) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (t *Team) AddContactType(id int, name string) int {
	answer := 0
	found := false
	next := 0
	sortid := -1
	for c, ctype := range t.ContactTypes {
		if next < ctype.Id {
			next = ctype.Id
		}
		if sortid < ctype.SortID {
			sortid = ctype.SortID
		}
		if ctype.Id == id {
			found = true
			ctype.Name = name
			t.ContactTypes[c] = ctype
			answer = ctype.Id
		}
	}
	if !found {
		ctype := &ContactType{
			Id:     next + 1,
			Name:   name,
			SortID: sortid + 1,
		}
		answer = ctype.Id
		t.ContactTypes = append(t.ContactTypes, *ctype)
	}
	sort.Sort(ByContactType(t.ContactTypes))
	return answer
}

func (t *Team) UpdateContactTypeSort(id int, direction string) {
	pos := -1
	for c, ctype := range t.ContactTypes {
		if ctype.Id == id {
			pos = c
		}
	}
	if pos >= 0 {
		oldSort := t.ContactTypes[pos].SortID
		newSort := -1
		otherpos := -1
		if strings.EqualFold(direction[:1], "u") && pos > 0 {
			newSort = t.ContactTypes[pos-1].SortID
			otherpos = pos - 1
		} else if strings.EqualFold(direction[:1], "d") && pos < len(t.ContactTypes)-1 {
			newSort = t.ContactTypes[pos+1].SortID
			otherpos = pos + 1
		}
		if newSort >= 0 {
			t.ContactTypes[pos].SortID = newSort
			t.ContactTypes[otherpos].SortID = oldSort
		}
	}
	sort.Sort(ByContactType(t.ContactTypes))
}

func (t *Team) DeleteContactType(id int) {
	pos := -1
	for c, ctype := range t.ContactTypes {
		if ctype.Id == id {
			pos = c
		}
	}
	if pos >= 0 {
		t.ContactTypes = append(t.ContactTypes[:pos], t.ContactTypes[pos+1:]...)
	}
	sort.Sort(ByContactType(t.ContactTypes))
	for c, cType := range t.ContactTypes {
		cType.SortID = c
		t.ContactTypes[c] = cType
	}
}

func (t *Team) AddSpecialtyType(id int, name string) int {
	answer := 0
	found := false
	next := 0
	sortid := -1
	for c, ctype := range t.SpecialtyTypes {
		if next < ctype.Id {
			next = ctype.Id
		}
		if sortid < ctype.SortID {
			sortid = ctype.SortID
		}
		if ctype.Id == id {
			found = true
			ctype.Name = name
			answer = ctype.Id
			t.SpecialtyTypes[c] = ctype
		}
	}
	if !found {
		ctype := &SpecialtyType{
			Id:     next + 1,
			Name:   name,
			SortID: sortid + 1,
		}
		answer = ctype.Id
		t.SpecialtyTypes = append(t.SpecialtyTypes, *ctype)
	}
	sort.Sort(ByContactType(t.ContactTypes))
	return answer
}

func (t *Team) UpdateSpecialtyTypeSort(id int, direction string) {
	pos := -1
	for c, ctype := range t.SpecialtyTypes {
		if ctype.Id == id {
			pos = c
		}
	}
	if pos >= 0 {
		oldSort := t.SpecialtyTypes[pos].SortID
		newSort := -1
		otherpos := -1
		if strings.EqualFold(direction[:1], "u") && pos > 0 {
			newSort = t.SpecialtyTypes[pos-1].SortID
			otherpos = pos - 1
		} else if strings.EqualFold(direction[:1], "d") && pos < len(t.SpecialtyTypes)-1 {
			newSort = t.SpecialtyTypes[pos+1].SortID
			otherpos = pos + 1
		}
		if newSort >= 0 {
			t.SpecialtyTypes[pos].SortID = newSort
			t.SpecialtyTypes[otherpos].SortID = oldSort
		}
	}
	sort.Sort(BySpecialtyType(t.SpecialtyTypes))
}

func (t *Team) DeleteSpecialtyType(id int) {
	pos := -1
	for c, ctype := range t.SpecialtyTypes {
		if ctype.Id == id {
			pos = c
		}
	}
	if pos >= 0 {
		t.SpecialtyTypes = append(t.SpecialtyTypes[:pos], t.SpecialtyTypes[pos+1:]...)
	}
	sort.Sort(BySpecialtyType(t.SpecialtyTypes))
	for c, cType := range t.SpecialtyTypes {
		cType.SortID = c
		t.SpecialtyTypes[c] = cType
	}
}

func (t *Team) PurgeOldData(date time.Time) {
	for c, co := range t.Companies {
		co.Purge(date)
		t.Companies[c] = co
	}
}
