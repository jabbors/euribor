package main

import (
	"testing"
	"time"
)

func TestKey(t *testing.T) {
	th := threshold{Email: "foo@bar.com", Maturity: "1w"}
	if th.Key() != "gorates_foo@bar.com;1w" {
		t.Errorf("expected gorates_foo@bar.com;1w, got %s", th.Key())
	}
}

func TestNewThresholdFromKeyVal(t *testing.T) {
	th, err := newThresholdFromKeyVal("gorates_foo@bar.com;1w", 1.0)
	if err != nil {
		t.Errorf("did not expect a failure from valid input")
	}
	if th.Email != "foo@bar.com" {
		t.Errorf("expected foo@bar.com, got %s", th.Email)
	}
	if th.Maturity != "1w" {
		t.Errorf("expected 1w, got %s", th.Maturity)
	}
}

func TestExceeded(t *testing.T) {
	positiveSampleRates := []rate{
		{time.Date(2009, time.November, 10, 0, 0, 0, 0, time.UTC), 1.0},
		{time.Date(2009, time.November, 11, 0, 0, 0, 0, time.UTC), 1.1},
		{time.Date(2009, time.November, 12, 0, 0, 0, 0, time.UTC), 1.2},
		{time.Date(2009, time.November, 13, 0, 0, 0, 0, time.UTC), 1.3},
		{time.Date(2009, time.November, 14, 0, 0, 0, 0, time.UTC), 1.4},
		{time.Date(2009, time.November, 15, 0, 0, 0, 0, time.UTC), 1.5},
		{time.Date(2009, time.November, 16, 0, 0, 0, 0, time.UTC), 1.6},
	}
	negativeSampleRates := []rate{
		{time.Date(2009, time.December, 10, 0, 0, 0, 0, time.UTC), -1.6},
		{time.Date(2009, time.December, 11, 0, 0, 0, 0, time.UTC), -1.5},
		{time.Date(2009, time.December, 12, 0, 0, 0, 0, time.UTC), -1.4},
		{time.Date(2009, time.December, 13, 0, 0, 0, 0, time.UTC), -1.3},
		{time.Date(2009, time.December, 14, 0, 0, 0, 0, time.UTC), -1.2},
		{time.Date(2009, time.December, 15, 0, 0, 0, 0, time.UTC), -1.1},
		{time.Date(2009, time.December, 16, 0, 0, 0, 0, time.UTC), -1.0},
	}
	testCases := []struct {
		limit    float64
		rates    []rate
		exceeded bool
	}{
		{2.0, positiveSampleRates, false},
		{1.5, positiveSampleRates, false},
		{1.3, positiveSampleRates, false},
		{1.2, positiveSampleRates, true},
		{1.1, positiveSampleRates, true},
		{-2.0, negativeSampleRates, true},
		{-1.5, negativeSampleRates, true},
		{-1.4, negativeSampleRates, true},
		{-1.3, negativeSampleRates, false},
		{-1.2, negativeSampleRates, false},
	}

	for _, tc := range testCases {
		th := threshold{Limit: tc.limit}
		exceeded := th.Exceeded(tc.rates)
		if exceeded != tc.exceeded {
			t.Errorf("test case with limit %v and input %v failed, expected %t got %t", tc.limit, tc.rates, tc.exceeded, exceeded)
		}
	}
}
