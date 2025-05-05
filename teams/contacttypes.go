package teams

type ContactType struct {
	Id     int    `json:"id" bson:"id"`
	Name   string `json:"name" bson:"name"`
	SortID int    `json:"sort" bson:"sort"`
}

type ByContactType []ContactType

func (c ByContactType) Len() int { return len(c) }
func (c ByContactType) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByContactType) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type SpecialtyType struct {
	Id     int    `json:"id" bson:"id"`
	Name   string `json:"name" bson:"name"`
	SortID int    `json:"sort" bson:"sort"`
}

type BySpecialtyType []SpecialtyType

func (c BySpecialtyType) Len() int { return len(c) }
func (c BySpecialtyType) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c BySpecialtyType) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
