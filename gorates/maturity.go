package main

var (
	maturities = []string{"1w", "1m", "3m", "6m", "12m"}
)

func isValidMaturity(m string) bool {
	switch m {
	case "1w":
		return true
	case "1m":
		return true
	case "3m":
		return true
	case "6m":
		return true
	case "12m":
		return true
	}
	return false
}
