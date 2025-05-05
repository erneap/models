package entries

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UsersSoapEntriesYear struct {
	ID      primitive.ObjectID `json:"id" bson:"_id"`
	User    primitive.ObjectID `json:"user" bson:"user"`
	Year    int                `json:"year" bson:"year"`
	Entries []SoapEntry        `json:"entries,omitempty" bson:"entries,omitempty"`
}
type BySoapUserYear []UsersSoapEntriesYear

func (c BySoapUserYear) Len() int { return len(c) }
func (c BySoapUserYear) Less(i, j int) bool {
	if c[i].User.Hex() == c[j].User.Hex() {
		return c[i].Year < c[j].Year
	}
	return c[i].User.Hex() < c[j].User.Hex()
}
func (c BySoapUserYear) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (u *UsersSoapEntriesYear) GetEntry(date time.Time) (*SoapEntry, error) {
	var entry *SoapEntry
	for _, e := range u.Entries {
		if e.IsEntry(date) {
			entry = &e
		}
	}
	if entry == nil {
		return nil, errors.New("not found")
	}
	return entry, nil
}

func (u *UsersSoapEntriesYear) AddEntry(entry SoapEntry) error {
	found := false
	for _, e := range u.Entries {
		if e.IsEntry(entry.Date) {
			found = true
		}
	}
	if found {
		return errors.New("already exists")
	}
	u.Entries = append(u.Entries, entry)
	sort.Sort(BySoapEntry(u.Entries))
	return nil
}

func (u *UsersSoapEntriesYear) ModifyEntry(date time.Time, field,
	value string) (*SoapEntry, error) {
	for e, entry := range u.Entries {
		if entry.IsEntry(date) {
			switch strings.ToLower(field) {
			case "scripturebook", "book":
				entry.Scripture.Book = value
			case "scripturechapter", "chapter":
				chptr, err := strconv.Atoi(value)
				if err != nil {
					return nil, err
				}
				entry.Scripture.Chapter = chptr
			case "scriptureverses", "verses", "verse":
				entry.Scripture.Verses = value
			case "scripturetext", "scripture":
				entry.Scripture.Text = value
			case "observations":
				entry.Observations = value
			case "application":
				entry.Application = value
			case "prayer":
				entry.Prayer.Text = value
			case "shareprayer", "share":
				entry.Prayer.Share = strings.EqualFold(value, "true")
			}
			u.Entries[e] = entry
			return &entry, nil
		}
	}
	return nil, errors.New("not found")
}

func (u *UsersSoapEntriesYear) RemoveEntry(date time.Time) error {
	pos := -1
	for e, entry := range u.Entries {
		if entry.IsEntry(date) {
			pos = e
		}
	}
	if pos < 0 {
		return errors.New("not found")
	}
	u.Entries = append(u.Entries[:pos], u.Entries[pos+1:]...)
	return nil
}
