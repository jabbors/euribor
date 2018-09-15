package main

var (
	retentions   = []string{"week", "month", "three_months", "six_months", "year", "two_years", "six_years"}
	retentionMap = map[string]string{
		"last-week":       "week",
		"last-month":      "month",
		"last-quarter":    "three_months",
		"last-six-months": "six_months",
		"last-year":       "year",
		"last-two-years":  "two_years",
		"last-six-years":  "six_years",
	}
)

func isValidRetention(r string) bool {
	if _, ok := retentionMap[r]; ok {
		return true
	}
	return false
}
