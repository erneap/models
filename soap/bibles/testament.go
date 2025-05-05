package bibles

type Testament struct {
	Code  string      `json:"code" bson:"code"`
	Title string      `json:"title" bson:"title"`
	Books []BibleBook `json:"books" bson:"_"`
}

type ByTestament []Testament

func (c ByTestament) Len() int { return len(c) }
func (c ByTestament) Less(i, j int) bool {
	return c[i].Code > c[j].Code
}
func (c ByTestament) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
