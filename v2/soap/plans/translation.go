package plans

type Translation struct {
	ID    uint   `bson:"_id" json:"id,omitempty"`
	Short string `bson:"short" json:"short"`
	Long  string `bson:"long" json:"long"`
}

type ByTranslation []Translation

func (c ByTranslation) Len() int { return len(c) }
func (c ByTranslation) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByTranslation) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type TranslationList struct {
	Translations []Translation `json:"list,omitempty"`
}
