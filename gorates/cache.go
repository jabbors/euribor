package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

var (
	historyCache map[string][]rate
)

// runs in go routine and takes care of refreshing the cache
func refreshCache() {
	historyCache = make(map[string][]rate)

	var lastRefresh time.Time
	var lastModified time.Time
	for {
		for _, maturity := range maturities {
			filename := fmt.Sprintf("%s/euribor-rates-%s.csv", historyPath, maturity)
			fi, err := os.Stat(filename)
			if err != nil {
				fmt.Println("[cache]: error get stat", err)
			}
			if fi.ModTime().After(lastModified) {
				lastModified = fi.ModTime()
			}
		}
		if !lastModified.After(lastRefresh) {
			time.Sleep(time.Minute)
			continue
		}

		// refresh
		log.Println("[cache]: refreshing history cache")
		for _, maturity := range maturities {
			file := fmt.Sprintf("%s/euribor-rates-%s.csv", historyPath, maturity)
			historyCache[maturity] = parseFile(file)
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
