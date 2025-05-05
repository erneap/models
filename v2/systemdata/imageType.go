package systemdata

type ImageType struct {
	ID           string      `json:"id" bson:"id"`
	Collected    uint        `json:"collected,omitempty" bson:"collected,omitempty"`
	NotCollected uint        `json:"notcollected,omitempty" bson:"notcollected,omitempty"`
	SortID       uint        `json:"sortID" bson:"sortID"`
	Subtypes     []ImageType `json:"subtypes,omitempty" bson:"subtypes,omitempty"`
}

type ByImageType []ImageType

func (c ByImageType) Len() int { return len(c) }
func (c ByImageType) Less(i, j int) bool {
	return c[i].SortID < c[j].SortID
}
func (c ByImageType) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
