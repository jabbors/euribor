package main

import (
	"fmt"
	"log"
)

// monitors user requested rate limits and sends out alerts when exceeded
func monitorRates() {
	log.Println("[monitor]: loading thresholds from redis")
	filter := ""
	thresholds := loadThresholds(filter)

	log.Println("[monitor]: processing thresholds...")
	for _, th := range thresholds {
		if th.Triggerd {
			continue
		}

		if th.Exceeded(historyCache[th.Maturity]) {
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
