package plans

import (
	"errors"
	"sort"
	"strconv"
	"strings"
)

type ReadingPeriod struct {
	ID   int          `json:"id" bson:"id"`
	Days []ReadingDay `json:"days,omitempty" bson:"days,omitempty"`
}

type ByReadingPeriod []ReadingPeriod

func (c ByReadingPeriod) Len() int { return len(c) }
func (c ByReadingPeriod) Less(i, j int) bool {
	return c[i].ID < c[j].ID
}
func (c ByReadingPeriod) Swap(i, j int) { c[i], c[j] = c[j], c[i] }

func (m *ReadingPeriod) IsCompleted() bool {
	answer := true
	for _, day := range m.Days {
		if !day.Completed {
			answer = false
		}
	}
	return answer
}

func (m *ReadingPeriod) AddReadingDay(day, bookid int, book string,
	chapter, start, end int) error {
	if day == 0 {
		return errors.New("day can't be zero")
	}
	found := false
	for d, rday := range m.Days {
		if rday.Day == day {
			found = true
			err := rday.AddPassage(bookid, book, chapter, start, end)
			if err != nil {
				return err
			}
			m.Days[d] = rday
		}
	}
	if !found {
		rday := &ReadingDay{
			Day: day,
		}
		if bookid > 0 {
			err := rday.AddPassage(bookid, book, chapter, start, end)
			if err != nil {
				return err
			}
		}
		m.Days = append(m.Days, *rday)
		sort.Sort(ByReadingDay(m.Days))
	}
	return nil
}

func (m *ReadingPeriod) UpdateReadingDay(day, id int, field, value string) error {
	if day == 0 {
		return errors.New("day can't be zero")
	}
	found := false
	sort.Sort(ByReadingDay(m.Days))
	for d, rday := range m.Days {
		if rday.Day == day {
			found = true
			switch strings.ToLower(field) {
			case "sortdays":
				if strings.ToLower(value) == "up" && d > 0 {
					temp := rday.Day
					rday.Day = m.Days[d-1].Day
					m.Days[d-1].Day = temp
				} else if strings.ToLower(value) == "down" && d < len(m.Days)-1 {
					temp := rday.Day
					rday.Day = m.Days[d+1].Day
					m.Days[d+1].Day = temp
				}
			case "addpassage":
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
				rday.AddPassage(id, book, chapter, start, end)
			case "deletepassage":
				rday.DeletePassage(id)
			default:
				err := rday.UpdatePassage(id, field, value)
				if err != nil {
					return err
				}
			}
			m.Days[d] = rday
		}
	}
	if !found {
		return errors.New("plan day not found")
	}
	return nil
}

func (m *ReadingPeriod) DeleteReadingDay(day int) error {
	if day == 0 {
		return errors.New("day can't be zero")
	}
	pos := -1
	for d, rday := range m.Days {
		if rday.Day == day {
			pos = d
		}
	}
	if pos < 0 {
		return errors.New("plan day not found")
	}
	m.Days = append(m.Days[:pos], m.Days[pos+1:]...)
	return nil
}

func (m *ReadingPeriod) ResetPeriod() {
	for d, rday := range m.Days {
		rday.ResetDay()
		m.Days[d] = rday
	}
}
