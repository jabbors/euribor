package main

import (
	"encoding/json"
	"strconv"
	"time"
)

type rate struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// MarshalJSON overrides marshal, prints shorter date string
func (r *rate) MarshalJSON() ([]byte, error) {
	type alias rate
	return json.Marshal(&struct {
		Date string `json:"date"`
		*alias
	}{
		Date:  r.Date.Format("2006-01-02"),
		alias: (*alias)(r),
	})
}

type rateSlice []rate

func (rs rateSlice) Values() string {
	str := "["
	for i, r := range rs {
		str += "["
		str += strconv.FormatInt(r.Date.Unix()*1000, 10) // in milli seconds
		str += ","
		str += strconv.FormatFloat(r.Value, 'g', 4, 64)
		str += "]"
		if i != len(rs)-1 {
			str += ","
		}
	}
	str += "]"

	return str
}
