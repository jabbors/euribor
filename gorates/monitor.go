package main

import (
	"fmt"
	"log"
)

// monitorService watches rates and sends out alerts when thresholds are exceeded
type monitorService struct {
	pushbulletToken string
	monitorCh       <-chan bool
}

func NewMonitorService(pushbulletToken string, monitorCh <-chan bool) *monitorService {
	ms := monitorService{
		pushbulletToken: pushbulletToken,
		monitorCh:       monitorCh,
	}

	ms.start()
	return &ms
}

func (ms *monitorService) start() {
	go ms.monitorRates()
}

// monitors user requested rate limits and sends out alerts when exceeded
func (ms *monitorService) monitorRates() {
	for {
		log.Println("[monitor]: waiting for signal")
		_ = <-ms.monitorCh

		log.Println("[monitor]: loading thresholds from redis")
		filter := ""
		thresholds := loadThresholds(filter)

		log.Println("[monitor]: processing thresholds...")
		for _, th := range thresholds {
			if th.Triggerd {
				continue
			}

			if th.Exceeded(historyCache[th.Maturity]) {
				err := th.Alert(ms.pushbulletToken)
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
