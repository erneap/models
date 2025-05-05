package plans

import (
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ReadingPlan struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Title     string             `json:"title,omitempty" bson:"title,omitempty"`
	UserID    string             `json:"userid,omitempty" bson:"userid,omitempty"`
	StartDate *time.Time         `json:"start,omitempty" bson:"start,omitempty"`
	Periods   []ReadingPeriod    `json:"periods,omitempty" bson:"periods,omitempty"`
}

type ByReadingPlan []ReadingPlan

func (c ByReadingPlan) Len() int { return len(c) }
func (c ByReadingPlan) Less(i, j int) bool {
	if c[i].UserID == c[j].UserID {
		return c[i].StartDate.Before(*c[j].StartDate)
	}
	return c[i].UserID < c[j].UserID
}
func (c ByReadingPlan) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (p *ReadingPlan) IsCompleted() bool {
	answer := true
	for _, month := range p.Periods {
		if !month.IsCompleted() {
			answer = false
		}
	}
	return answer
}

func (p *ReadingPlan) AddPeriod(days int) error {
	prd := &ReadingPeriod{
		ID: len(p.Periods) + 1,
	}
	for d := 0; d < days; d++ {
		err := prd.AddReadingDay(d+1, 0, "", 0, 0, 0)
		if err != nil {
			return err
		}
	}
	p.Periods = append(p.Periods, *prd)
	return nil
}

func (p *ReadingPlan) UpdatePeriod(prdid, day, id int, field, value string) error {
	found := false
	sort.Sort(ByReadingPeriod(p.Periods))
	for i, prd := range p.Periods {
		if prd.ID == prdid {
			found = true
			switch strings.ToLower(field) {
			case "sortperiods":
				if strings.ToLower(value) == "up" && i > 0 {
					temp := prd.ID
					prd.ID = p.Periods[i-1].ID
					p.Periods[i-1].ID = temp
				} else if strings.ToLower(value) == "down" && i < len(p.Periods)-1 {
					temp := prd.ID
					prd.ID = p.Periods[i+1].ID
					p.Periods[i+1].ID = temp
				}
			case "addday":
				parts := strings.Split(value, "|")
				book := ""
				chapter := 0
				start := 0
				end := 0
				if len(parts) > 0 {
					book = parts[0]
				}
				if len(parts) > 1 {
					chapter, _ = strconv.Atoi(parts[1])
				}
				if len(parts) > 3 {
					start, _ = strconv.Atoi(parts[2])
					end, _ = strconv.Atoi(parts[3])
				}
				prd.AddReadingDay(day, id, book, chapter, start, end)
			case "deleteday":
				prd.DeleteReadingDay(day)
			default:
				err := prd.UpdateReadingDay(day, id, field, value)
				if err != nil {
					return err
				}
			}
			p.Periods[i] = prd
		}
	}
	if !found {
		return errors.New("reading plan not found")
	}
	return nil
}

func (p *ReadingPlan) DeletePeriod(id int) error {
	pos := -1
	for i, prd := range p.Periods {
		if prd.ID == id {
			pos = i
		}
	}
	if pos < 0 {
		return errors.New("period not found")
	}
	p.Periods = append(p.Periods[:pos], p.Periods[pos+1:]...)
	sort.Sort(ByReadingPeriod(p.Periods))
	for i, prd := range p.Periods {
		prd.ID = i + 1
		p.Periods[i] = prd
	}
	return nil
}

func (p *ReadingPlan) ResetReadingPlan() {
	for m, prd := range p.Periods {
		prd.ResetPeriod()
		p.Periods[m] = prd
	}
}
