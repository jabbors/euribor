package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

type threshold struct {
	Email    string
	Limit    float64
	Maturity string
	Triggerd bool
}

func (t threshold) Key() string {
	return fmt.Sprintf("gorates_%s;%s", t.Email, t.Maturity)
}

func newThreshold(email string, limit float64, maturity string) threshold {
	return threshold{Email: email, Limit: limit, Maturity: maturity}
}

func newThresholdFromKeyVal(key string, value float64) (threshold, error) {
	key = strings.TrimLeft(key, "gorates_")
	parts := strings.Split(key, ";")
	if len(parts) != 2 {
		return threshold{}, fmt.Errorf("email and maturity not found in key")
	}

	return threshold{Email: parts[0], Maturity: parts[1], Limit: value}, nil
}

// runs in go routine and takes care of monitoring rates and triggering alerts
func monitorRates() {
	historyRates := make(map[string][]rate)
	for {
		if time.Since(lastRefresh) < time.Hour*24 {
			time.Sleep(time.Minute)
			continue
		}

		log.Println("[monitor]: loading rates from files")
		for _, maturity := range maturities {
			// TODO: make path to files configureable
			file := fmt.Sprintf("%s/euribor-rates-%s.csv", historyPath, maturity)
			historyRates[maturity] = parseFile(file)
		}

		log.Println("[monitor]: loading thresholds from redis")
		filter := ""
		thresholds := loadThresholds(filter)

		log.Println("[monitor]: monitoring rates...")
		for _, th := range thresholds {
			if th.Triggerd {
				continue
			}

			if thresholdExceeded(th.Limit, historyRates[th.Maturity]) {
				err := sendThresholdAlert(th)
				if err != nil {
					fmt.Printf("[monitor]: error: failed sending alert for threshold %s: %v\n", th.Key(), err)
					continue
				}
				err = removeThreshold(th)
				if err != nil {
					fmt.Printf("[monitor]: error: failed removing threshold %s: %v\n", th.Key(), err)
				}
			}
		}
		log.Println("[monitor]: monitoring completed")
	}
}

func thresholdExceeded(threshold float64, rates []rate) bool {
	// we need at least 5 samples to determine a trend
	if len(rates) < 5 {
		return false
	}
	for i := len(rates) - 1; i > len(rates)-6; i-- {
		if rates[i].Value >= threshold {
			continue
		}
		return false
	}
	return true
}
