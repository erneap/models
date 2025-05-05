package bibles

import (
	"strings"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BibleBook struct {
	Id        primitive.ObjectID `json:"-" bson:"_id"`
	BookId    int                `json:"id" bson:"bookid"`
	Version   string             `json:"-" bson:"version"`
	Testament string             `json:"-" bson:"testament"`
	Code      string             `json:"code,omitempty" bson:"code,omitempty"`
	Title     string             `json:"title" bson:"title"`
	Chapters  []BibleChapter     `json:"chapters,omitempty" bson:"chapter,omitempty"`
}

type ByBibleBook []BibleBook

func (c ByBibleBook) Len() int { return len(c) }
func (c ByBibleBook) Less(i, j int) bool {
	return c[i].BookId < c[j].BookId
}
func (c ByBibleBook) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (b *BibleBook) AddChapter() (*BibleChapter, error) {
	chptr := &BibleChapter{
		Id: len(b.Chapters) + 1,
	}
	b.Chapters = append(b.Chapters, *chptr)
	return chptr, nil
}

func (b *BibleBook) IsBook(name string) bool {
	length := len(name)
	return strings.EqualFold(b.Title[:length], name)
}

type StandardBibleBook struct {
	ID         int                    `json:"id" bson:"id"` // used for identification and sort order
	Title      string                 `json:"title" bson:"title"`
	Apocryphal bool                   `json:"aprocryphal" bson:"aprocryphal"`
	Chapters   []StandardBibleChapter `json:"chapters,omitempty" bson:"chapters,omitempty"`
}
type ByStandardBibleBook []StandardBibleBook

func (c ByStandardBibleBook) Len() int { return len(c) }
func (c ByStandardBibleBook) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByStandardBibleBook) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
