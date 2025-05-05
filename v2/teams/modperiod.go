package teams

import "time"

type ModPeriod struct {
	Year  int       `json:"year" bson:"year"`
	Start time.Time `json:"start" bson:"start"`
	End   time.Time `json:"end" bson:"end"`
}

type ByModPeriod []ModPeriod

func (c ByModPeriod) Len() int { return len(c) }
func (c ByModPeriod) Less(i, j int) bool {
	return c[i].Year < c[j].Year
}
func (c ByModPeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
