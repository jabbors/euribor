package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
)

const timeFormat = "2006-01-02"

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

var lastRefresh time.Time
var maturities = []string{"1w", "2w", "1m", "2m", "3m", "6m", "9m", "12m"}
var retentions = []string{"week", "month", "three_months", "six_months", "year", "two_years", "six_years"}
var historyPath string
var historyCache map[string][]rate
var influxCache map[string]map[string][]rate

// runs in go routine and takes care of refreshing the cache
func refreshCache() {
	historyCache = make(map[string][]rate)
	influxCache = make(map[string]map[string][]rate)

	for {
		if time.Since(lastRefresh) < time.Hour*24 {
			time.Sleep(time.Minute)
			continue
		}

		// refresh
		log.Println("refreshing history cache")
		for _, maturity := range maturities {
			// TODO: make path to files configureable
			file := fmt.Sprintf("%s/euribor-rates-%s.csv", historyPath, maturity)
			historyCache[maturity] = parseFile(file)
		}
		log.Println("refreshing influx cache")
		for _, retention := range retentions {
			results := queryInflux(retention)
			cache := make(map[string][]rate)
			// fmt.Println(results.Values)
			// fmt.Println(reflect.TypeOf(results.Values).Kind())
			for _, value := range results.Values {
				// fmt.Println(value, reflect.TypeOf(value).Kind(), reflect.TypeOf(value))
				m, r, err := transformInfluxValueToRate(reflect.ValueOf(value).Interface().([]interface{}))
				if err != nil {
					log.Println("error converting influx value to rate")
					continue
				}
				if rates, ok := cache[m]; ok {
					rates = append(rates, r)
					cache[m] = rates
				} else {
					cache[m] = []rate{r}
				}
			}
			influxCache[retention] = cache
		}
		log.Println("cache refresh completed")

		lastRefresh = time.Now()
	}
}

func parseFile(file string) []rate {
	f, err := os.Open(file)
	if err != nil {
		log.Println(err)
		return []rate{}
	}

	r := csv.NewReader(f)
	r.Comment = '#'
	r.FieldsPerRecord = 2

	records, err := r.ReadAll()
	if err != nil {
		log.Println(err)
		return []rate{}
	}

	rates := []rate{}
	for _, record := range records {
		date, err := time.ParseInLocation(timeFormat, record[0], time.UTC)
		if err != nil {
			log.Println(err)
			continue
		}
		value, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			log.Println(err)
			continue
		}
		rates = append(rates, rate{date, value})
	}

	return rates
}

func isValidRetention(r string) bool {
	// TODO: fix some kind of mapping
	// last-week
	// last-month
	// last-quater
	// last-six-months
	// last-year
	// last-two-years
	// last-six-years

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

// index handler
func index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	log.Printf("%v /", r.RemoteAddr)
	fmt.Fprint(w, "Welcome to the Euribor rates service!\n")
}

// influx handler
func influx(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Printf("%v /rates/app/%s/%s", r.RemoteAddr, params.ByName("retention"), params.ByName("maturity"))
	retention := params.ByName("retention")
	if isValidRetention(retention) == false {
		//TODO: return http error code
		fmt.Fprintf(w, errorMsg("uknown retention"))
		return
	}

	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		// TODO: return http error code
		fmt.Fprintf(w, errorMsg("uknown maturity"))
		return
	}

	rates := []rate{}
	for _, r := range influxCache[retention][maturity] {
		rates = append(rates, r)
	}

	jsonData, err := json.Marshal(rates)
	if err != nil {
		fmt.Fprintf(w, errorMsg(err.Error()))
		return
	}

	fmt.Fprintf(w, string(jsonData))
}

// highstock handler
func highstock(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Printf("%v /rates/app/hs/%s", r.RemoteAddr, params.ByName("maturity"))
	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		// TODO: return http error code
		fmt.Fprintf(w, errorMsg("uknown maturity"))
		return
	}

	rates := rateSlice{}
	for _, r := range historyCache[maturity] {
		rates = append(rates, r)
	}

	fmt.Fprintf(w, rates.Values())
}

// history handler
func history(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Printf("%v /rates/history/%s/%s", r.RemoteAddr, params.ByName("year"), params.ByName("maturity"))
	year, err := strconv.ParseInt(params.ByName("year"), 10, 32)
	if err != nil {
		// TODO: return http error code
		fmt.Fprint(w, errorMsg(err.Error()))
		return
	}
	if year < 2010 || year > int64(time.Now().Year()) {
		fmt.Fprintf(w, errorMsg("no data"))
		return
	}

	maturity := params.ByName("maturity")
	if isValidMaturity(maturity) == false {
		// TODO: return http error code
		fmt.Fprintf(w, errorMsg("uknown maturity"))
		return
	}

	rates := []rate{}
	for _, r := range historyCache[maturity] {
		if int64(r.Date.Year()) != year {
			continue
		}
		rates = append(rates, r)
	}

	jsonData, err := json.Marshal(rates)
	if err != nil {
		fmt.Fprintf(w, errorMsg(err.Error()))
		return
	}

	fmt.Fprintf(w, string(jsonData))
}

// webapp handler
func webapp(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	log.Printf("%v /webapp", r.RemoteAddr)
	url := baseURL(r)

	fmt.Fprintf(w, renderWebapp(url))
}

func errorMsg(msg string) string {
	return fmt.Sprintf("{\"error\":\"%s\"}", msg)
}

func main() {
	var host string
	var port string
	flag.StringVar(&host, "host", "localhost", "host to bind to")
	flag.StringVar(&port, "port", "8080", "port to bind to")
	flag.StringVar(&historyPath, "history-path", ".", "path to history rate CSV files")
	flag.Parse()

	go refreshCache()

	router := httprouter.New()
	router.RedirectTrailingSlash = true

	// routes for general info
	// TODO: add routes for list of supported retentions/maturities
	router.GET("/", index)

	// routes to serve the app
	router.GET("/rates/app/if/:retention/:maturity", influx)
	router.GET("/rates/app/hs/:maturity", highstock)

	// routes to serve historical queries
	router.GET("/rates/history/:year/:maturity", history)

	// routes to serve the webapp
	router.GET("/webapp", webapp)

	log.Printf("listening on %s:%s", host, port)
	log.Fatal(http.ListenAndServe(host+":"+port, router))
}
