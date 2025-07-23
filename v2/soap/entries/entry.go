package entries

import (
	"sort"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Entry struct {
	EntryDate    time.Time `bson:"entrydate" json:"entryDate"`
	Title        string    `bson:"title" json:"title"`
	Scripture    string    `bson:"scripture" json:"scripture"`
	Observations string    `bson:"observations" json:"observations"`
	Application  string    `bson:"application" json:"application"`
	Prayer       string    `bson:"prayer" json:"prayer"`
}

type ByEntries []Entry

func (c ByEntries) Len() int { return len(c) }
func (c ByEntries) Less(i, j int) bool {
	return c[i].EntryDate.Before(c[j].EntryDate)
}
func (c ByEntries) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type EntryList struct {
	ID      primitive.ObjectID `bson:"_id" json:"id"`
	UserID  primitive.ObjectID `bson:"userid" json:"userid"`
	Name    string             `bson:"name" json:"name"`
	Year    uint               `bson:"year" json:"year"`
	Entries []Entry            `bson:"entries,omitempty" json:"entries,omitempty"`
}

type ByEntryList []EntryList

func (c ByEntryList) Len() int { return len(c) }
func (c ByEntryList) Less(i, j int) bool {
	if strings.EqualFold(c[i].Name, c[j].Name) {
		return c[i].Year < c[j].Year
	}
	return c[i].Name < c[j].Name
}
func (c ByEntryList) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (el *EntryList) AddEntry(entry Entry) {
	found := false
	for i, e := range el.Entries {
		if e.EntryDate.Year() == entry.EntryDate.Year() &&
			e.EntryDate.Month() == entry.EntryDate.Month() &&
			e.EntryDate.Day() == entry.EntryDate.Day() {
			found = true
			e.Title = entry.Title
			e.Scripture = entry.Scripture
			e.Observations = entry.Observations
			e.Application = entry.Application
			e.Prayer = entry.Prayer
			el.Entries[i] = e
		}
	}
	if !found {
		el.Entries = append(el.Entries, entry)
		sort.Sort(ByEntries(el.Entries))
	}
}

func (el *EntryList) UpdateEntry(date time.Time, field, value string) {
	for i, e := range el.Entries {
		if e.EntryDate.Year() == date.Year() &&
			e.EntryDate.Month() == date.Month() &&
			e.EntryDate.Day() == date.Day() {
			switch strings.ToLower(field) {
			case "title":
				e.Title = value
			case "scripture":
				e.Scripture = value
			case "observations":
				e.Observations = value
			case "application":
				e.Application = value
			case "prayer":
				e.Prayer = value
			}
			el.Entries[i] = e
		}
	}
}

func (el *EntryList) DeleteEntry(date time.Time) {
	found := -1
	for i, e := range el.Entries {
		if e.EntryDate.Year() == date.Year() &&
			e.EntryDate.Month() == date.Month() &&
			e.EntryDate.Day() == date.Day() {
			found = i
		}
	}
	if found >= 0 {
		el.Entries = append(el.Entries[:found], el.Entries[found+1:]...)
	}
}
