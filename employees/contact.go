package employees

type Contact struct {
	Id     int    `json:"id" bson:"id"`
	TypeID int    `json:"typeid" bson:"typeid"`
	SortID int    `json:"sort" bson:"sort"`
	Value  string `json:"value" bson:"value"`
}

type ByEmployeeContact []Contact

func (c ByEmployeeContact) Len() int { return len(c) }
func (c ByEmployeeContact) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByEmployeeContact) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type Specialty struct {
	Id          int  `json:"id" bson:"id"`
	SpecialtyID int  `json:"specialtyid" bson:"specialtyid"`
	SortID      int  `json:"sort" bson:"sort"`
	Qualified   bool `json:"qualified" bson:"qualified"`
}

type ByEmployeeSpecialty []Specialty

func (c ByEmployeeSpecialty) Len() int { return len(c) }
func (c ByEmployeeSpecialty) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByEmployeeSpecialty) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
