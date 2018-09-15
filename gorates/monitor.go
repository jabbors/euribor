package main

import (
	"fmt"
	"log"
	"time"
)

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

			if th.Exceeded(historyRates[th.Maturity]) {
				err := th.Alert()
				if err != nil {
					fmt.Printf("[monitor]: error: failed sending alert for threshold %s: %v\n", th.Key(), err)
					continue
				}
				err = th.Remove()
				if err != nil {
					fmt.Printf("[monitor]: error: failed removing threshold %s: %v\n", th.Key(), err)
				}
			}
		}
		log.Println("[monitor]: monitoring completed")
	}
}
