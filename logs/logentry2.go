package logs

import (
	"strings"
	"time"
)

type LogEntry2 struct {
	EntryDate time.Time `json:"entrydate" bson:"entrydate"`
	Category  string    `json:"category" bson:"category"`
	Title     string    `json:"title" bson:"title"`
	Message   string    `json:"message" bson:"message"`
	Name      string    `json:"requestor" bson:"requestor"`
}

func (le *LogEntry2) ToString() string {
	return le.EntryDate.UTC().Format("060102T150405Z") + "|" +
		le.Category + "|" + le.Title + "|" + le.Message +
		"|" + le.Name
}

func (le *LogEntry2) FromString(line string) {
	parts := strings.Split(line, "|")
	for i, part := range parts {
		switch i {
		case 0:
			le.EntryDate, _ = time.ParseInLocation("060102T150405Z", part, time.UTC)
		case 1:
			le.Category = part
		case 2:
			le.Title = part
		case 3:
			le.Message = part
		case 4:
			le.Name = part
		}
	}
}

type ByLogEntry2 []LogEntry2

func (c ByLogEntry2) Len() int { return len(c) }
func (c ByLogEntry2) Less(i, j int) bool {
	return c[i].EntryDate.Before(c[j].EntryDate)
}
func (c ByLogEntry2) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
