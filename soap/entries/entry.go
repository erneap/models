package entries

import (
	"time"
)

type SoapEntry struct {
	Date         time.Time          `json:"date" bson:"date"`
	Scripture    SoapScriptureEntry `json:"scripture" bson:"scripture"`
	Observations string             `json:"observations" bson:"observations"`
	Application  string             `json:"application" bson:"application"`
	Prayer       SoapPrayerEntry    `jsoon:"prayer" bson:"prayer"`
}
type BySoapEntry []SoapEntry

func (c BySoapEntry) Len() int { return len(c) }
func (c BySoapEntry) Less(i, j int) bool {
	return c[i].Date.Before(c[j].Date)
}
func (c BySoapEntry) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (e *SoapEntry) IsEntry(date time.Time) bool {
	return (e.Date.Year() == date.Year() && e.Date.Month() == date.Month() &&
		e.Date.Day() == date.Day())
}

type SoapScriptureEntry struct {
	Version string `json:"version" bson:"version"`
	Book    string `json:"book" bson:"book"`
	Chapter int    `json:"chapter" bson:"chapter"`
	Verses  string `json:"verses" bson:"verses"`
	Text    string `json:"text" bson:"text"`
}

type SoapPrayerEntry struct {
	Text  string `json:"text" bson:"text"`
	Share bool   `json:"share" bson:"share"`
}
