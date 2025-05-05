package plans

type Passage struct {
	ID         int    `json:"id,omitempty" bson:"id,omitempty"`
	BookID     int    `json:"bookid" bson:"bookid"`
	Book       string `json:"book" bson:"book"`
	Chapter    int    `json:"chapter" bson:"chapter"`
	StartVerse int    `json:"startverse,omitempty" bson:"startverse,omitempty"`
	EndVerse   int    `json:"endverse,omitempty" bson:"endverse,omitempty"`
	Passage    string `json:"passage,omitempty" bson:"passage,omitempty"`
	Completed  bool   `json:"completed,omitempty" bson:"completed"`
}

type ByPassage []Passage

func (c ByPassage) Len() int { return len(c) }
func (c ByPassage) Less(i, j int) bool {
	if c[i].BookID == c[j].BookID {
		if c[i].Chapter == c[j].Chapter {
			if c[i].StartVerse == c[j].StartVerse {
				return c[i].EndVerse < c[j].EndVerse
			}
			return c[i].StartVerse < c[j].StartVerse
		}
		return c[i].Chapter < c[j].Chapter
	}
	return c[i].BookID < c[j].BookID
}
func (c ByPassage) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p *Passage) ResetPassage() {
	p.Completed = false
}
