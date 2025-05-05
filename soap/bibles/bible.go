package bibles

import (
	"errors"
	"sort"
	"strings"

	"github.com/erneap/go-models/soap/plans"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Bible struct {
	Id         primitive.ObjectID `json:"id" bson:"_id"`
	Version    string             `json:"version,omitempty" bson:"version,omitempty"`
	Name       string             `json:"name,omitempty" bson:"name,omitempty"`
	Language   string             `json:"language,omitempty" bson:"language,omitempty"`
	Testaments []Testament        `json:"Testaments" bson:"Testaments"`
}

type ByBible []Bible

func (c ByBible) Len() int { return len(c) }
func (c ByBible) Less(i, j int) bool {
	return c[i].Version < c[j].Version
}
func (c ByBible) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (b *Bible) AddPassage(book string, chptr, start, end int,
	text string) *plans.Passage {
	var passage *plans.Passage
	if len(b.Testaments) == 0 {
		testament := &Testament{
			Code:  "ot",
			Title: "Old Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
		testament = &Testament{
			Code:  "nt",
			Title: "New Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
	}
	found := false
	for tid, testament := range b.Testaments {
		for bid, bk := range testament.Books {
			if strings.EqualFold(bk.Title, book) || strings.EqualFold(bk.Code, book) {
				for c, ch := range bk.Chapters {
					if ch.Id == chptr {
						for p, psg := range ch.Passages {
							if psg.StartVerse == start && psg.EndVerse == end {
								found = true
								psg.Passage = text
								ch.Passages[p] = psg
							}
						}
						if !found {
							psg := &plans.Passage{
								ID:         len(ch.Passages) + 1,
								BookID:     bk.BookId,
								Book:       book,
								Chapter:    chptr,
								StartVerse: start,
								EndVerse:   end,
								Passage:    text,
							}
							passage = psg
							found = true
							ch.Passages = append(ch.Passages, *psg)
							sort.Sort(plans.ByPassage(ch.Passages))
						}
						bk.Chapters[c] = ch
					}
				}
				if !found {
					// chapter and passage not found, so add chapter and passage at once
					ch := &BibleChapter{
						Id: chptr,
					}
					psg := &plans.Passage{
						ID:         len(ch.Passages) + 1,
						BookID:     bk.BookId,
						Book:       book,
						Chapter:    chptr,
						StartVerse: start,
						EndVerse:   end,
						Passage:    text,
					}
					passage = psg
					ch.Passages = append(ch.Passages, *psg)
					bk.Chapters = append(bk.Chapters, *ch)
					sort.Sort(ByBibleChapter(bk.Chapters))
					found = true
				}
				testament.Books[bid] = bk
			}
			b.Testaments[tid] = testament
		}
		if !found {
			bk := &BibleBook{
				BookId: len(testament.Books) + 1,
				Code:   strings.ToLower(book[:2]),
				Title:  book,
			}
			ch := &BibleChapter{
				Id: chptr,
			}
			psg := &plans.Passage{
				ID:         len(ch.Passages) + 1,
				BookID:     bk.BookId,
				Book:       book,
				Chapter:    chptr,
				StartVerse: start,
				EndVerse:   end,
				Passage:    text,
			}
			ch.Passages = append(ch.Passages, *psg)
			passage = psg
			bk.Chapters = append(bk.Chapters, *ch)
			testament.Books = append(testament.Books, *bk)
		}
	}
	return passage
}

func (b *Bible) GetPassageText(book string, chptr, start,
	end int) (string, error) {
	if len(b.Testaments) == 0 {
		testament := &Testament{
			Code:  "ot",
			Title: "Old Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
		testament = &Testament{
			Code:  "nt",
			Title: "New Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
	}
	answer := ""
	found := false
	for _, testament := range b.Testaments {
		for _, bk := range testament.Books {
			if strings.EqualFold(bk.Title, book) || strings.EqualFold(bk.Code, book) {
				for _, ch := range bk.Chapters {
					if ch.Id == chptr {
						if start == 0 && len(ch.Passages) > 0 {
							found = true
							answer = ch.Passages[0].Passage
						} else if start > 0 {
							for _, psg := range ch.Passages {
								if psg.StartVerse == start && psg.EndVerse == end {
									found = true
									answer = psg.Passage
								}
							}
						}
					}
				}
			}
		}
	}
	if !found || answer == "" {
		return "", errors.New("not Found")
	}
	return answer, nil
}

func (b *Bible) RemovePassage(book string, chptr, start,
	end int) (*plans.Passage, error) {
	if len(b.Testaments) == 0 {
		testament := &Testament{
			Code:  "ot",
			Title: "Old Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
		testament = &Testament{
			Code:  "nt",
			Title: "New Testament",
		}
		b.Testaments = append(b.Testaments, *testament)
	}
	var passage *plans.Passage
	for tid, testament := range b.Testaments {
		for i, bk := range testament.Books {
			if strings.EqualFold(bk.Title, book) || strings.EqualFold(bk.Code, book) {
				for c, ch := range bk.Chapters {
					if ch.Id == chptr {
						pos := -1
						for p, psg := range ch.Passages {
							if psg.StartVerse == start && psg.EndVerse == end {
								pos = p
								passage = &psg
							}
						}
						if pos >= 0 {
							ch.Passages = append(ch.Passages[:pos], ch.Passages[pos+1:]...)
						}
					}
					bk.Chapters[c] = ch
				}
				testament.Books[i] = bk
			}
		}
		b.Testaments[tid] = testament
	}
	if passage == nil {
		return nil, errors.New("not found")
	}
	return passage, nil
}

type BibleStandards struct {
	Books     []StandardBibleBook `json:"books,omitempty" bson:"books,omitempty"`
	Languages []BibleLanguage     `json:"languages,omitempty" bson:"languages,omitempty"`
}

type BibleLanguage struct {
	ID       primitive.ObjectID `json:"-" bson:"_id"`
	Code     string             `json:"code" bson:"code"`
	Title    string             `json:"title" bson:"title"`
	Versions []BibleVersion     `json:"versions,omitempty" bson:"-"`
	Bibles   []Bible            `json:"bibles,omitempty" bson:"-"`
}
type ByBibleLanguage []BibleLanguage

func (c ByBibleLanguage) Len() int { return len(c) }
func (c ByBibleLanguage) Less(i, j int) bool {
	if strings.EqualFold(c[i].Code, c[j].Code) {
		return strings.ToLower(c[i].Title) < strings.ToLower(c[j].Title)
	}
	return strings.ToLower(c[i].Code) < strings.ToLower(c[j].Code)
}
func (c ByBibleLanguage) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

type BibleVersion struct {
	Code  string `json:"code" bson:"code"`
	Title string `json:"title" bson:"title"`
}
type ByBibleVersion []BibleVersion

func (c ByBibleVersion) Len() int { return len(c) }
func (c ByBibleVersion) Less(i, j int) bool {
	return strings.ToLower(c[i].Title) < strings.ToLower(c[j].Title)
}
func (c ByBibleVersion) Swap(i, j int) { c[i], c[j] = c[j], c[i] }
