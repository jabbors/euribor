package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strconv"
	"time"
)

var (
	lastRefresh  time.Time
	historyCache map[string][]rate
	influxCache  map[string]map[string][]rate
)

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
		log.Println("[cache]: refreshing history cache")
		for _, maturity := range maturities {
			// TODO: make path to files configureable
			file := fmt.Sprintf("%s/euribor-rates-%s.csv", historyPath, maturity)
			historyCache[maturity] = parseFile(file)
		}
		log.Println("[cache]: refreshing influx cache")
		for _, retention := range retentions {
			results := queryInflux(retention)
			cache := make(map[string][]rate)
			// fmt.Println(results.Values)
			// fmt.Println(reflect.TypeOf(results.Values).Kind())
			for _, value := range results.Values {
				// fmt.Println(value, reflect.TypeOf(value).Kind(), reflect.TypeOf(value))
				m, r, err := transformInfluxValueToRate(reflect.ValueOf(value).Interface().([]interface{}))
				if err != nil {
					log.Println("[cache]: error converting influx value to rate")
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
		log.Println("[cache]: refresh completed")
		go monitorRates()

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
