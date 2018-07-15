package main

import "fmt"

func sendThresholdAlert(th threshold) error {
	fmt.Println("sending threshold alert to", th.Email)
	// TODO: implement sending alert through pushbullet API
	return nil
}
