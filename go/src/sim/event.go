package sim

import (
	"time"
)

type Event struct {
	Date time.Time `json:"date"`
	Value string `json:"value"`
	Severity int `json:"severity"`
}

type EventList []*Event

func (el *EventList) Add(date time.Time, value string, severity int) {
	*el = append(*el, &Event{Date: date, Value: value, Severity: severity})
}

func (el *EventList) Has(value string) bool {
	for _, e := range *el {
		if e.Value == value {
			return true
		}
	}
	return false
}

func (el *EventList) Len() int {
	return len(*el)
}

func (el *EventList) Swap(i, j int) {
	(*el)[i], (*el)[j] = (*el)[j], (*el)[i]
}

func (el *EventList) Less(i, j int) bool {
	return (*el)[i].Date.Before((*el)[j].Date)
}

