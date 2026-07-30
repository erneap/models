package labor

type Workcode struct {
	Id        string `json:"id" bson:"id"`
	Title     string `json:"title" bson:"title"`
	StartTime uint64 `json:"start" bson:"start"`
	ShiftCode string `json:"shiftCode" bson:"shiftCode"`
	AltCode   string `json:"altcode,omitempty" bson:"altcode,omitempty"`
	Search    string `json:"search,omitempty" bson:"search,omitempty"`
	IsLeave   bool   `json:"isLeave" bson:"isLeave"`
	TextColor string `json:"textcolor" bson:"textcolor"`
	BackColor string `json:"backcolor" bson:"backcolor"`
}

type ByWorkcode []Workcode

func (c ByWorkcode) Len() int { return len(c) }
func (c ByWorkcode) Less(i, j int) bool {
	if c[i].IsLeave == c[j].IsLeave {
		return c[i].Id < c[j].Id
	}
	if c[i].IsLeave && !c[j].IsLeave {
		return false
	}
	return true
}
func (c ByWorkcode) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
