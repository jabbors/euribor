package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

const (
	appName    = "gorates"
	timeFormat = "2006-01-02"
)

var version string

var maturities = []string{"1w", "2w", "1m", "2m", "3m", "6m", "9m", "12m"}
var retentions = []string{"week", "month", "three_months", "six_months", "year", "two_years", "six_years"}
var retentionMap = map[string]string{
	"last-week":       "week",
	"last-month":      "month",
	"last-quarter":    "three_months",
	"last-six-months": "six_months",
	"last-year":       "year",
	"last-two-years":  "two_years",
	"last-six-years":  "six_years",
}
var webRoot string
var historyPath string

func isValidRetention(r string) bool {
	if _, ok := retentionMap[r]; ok {
		return true
	}
	return false
}

func isValidMaturity(m string) bool {
	switch m {
	case "1w":
		return true
	case "2w":
		return true
	case "1m":
		return true
	case "2m":
		return true
	case "3m":
		return true
	case "6m":
		return true
	case "9m":
		return true
	case "12m":
		return true
	}
	return false
}

func main() {
	var host string
	var port string
	var verFlag bool
	flag.StringVar(&host, "host", "localhost", "host to bind to")
	flag.StringVar(&port, "port", "8080", "port to bind to")
	flag.StringVar(&webRoot, "web-root", "", "root path if hosted behind a proxy")
	flag.StringVar(&historyPath, "history-path", ".", "path to history rate CSV files")
	flag.BoolVar(&verFlag, "version", false, "print version and exit")
	flag.Parse()

	if verFlag {
		fmt.Printf("gorates: version=%s\n", version)
		return
	}

	go refreshCache()
	go monitorRates()

	router := httprouter.New()
	router.RedirectTrailingSlash = true

	// routes for general info
	// TODO: add routes for list of supported retentions/maturities
	router.GET("/", NoAuthHandler(indexHandler))
	router.GET("/version", NoAuthHandler(versionHandler))

	// routes to serve the app
	router.GET("/rates/app/if/:retention/:maturity", NoAuthHandler(influxHandler))
	router.GET("/rates/app/hs/:maturity", NoAuthHandler(highstockHandler))

	// routes to serve historical queries
	router.GET("/rates/history/:year/:maturity", NoAuthHandler(historyHandler))

	// routes to serve the webapp
	router.GET("/webapp", NoAuthHandler(webappHandler))

	// routes to manage alerts
	router.PUT("/alert/:email/:maturity/:limit", NoAuthHandler(alertAddHandler))
	router.DELETE("/alert/:email/:maturity/:limit", NoAuthHandler(alertRemoveHandler))
	router.GET("/alert/:email", NoAuthHandler(alertListHandler))

	log.Printf("listening on %s:%s", host, port)
	log.Fatal(http.ListenAndServe(host+":"+port, router))
}
