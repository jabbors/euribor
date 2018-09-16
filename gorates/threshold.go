package main

import (
	"fmt"
	"strings"
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

func (t threshold) Add() error {
	client, err := getConnection()
	if err != nil {
		fmt.Println("error: failed connecting to redis:", err)
		return err
	}
	err = client.Set(t.Key(), t.Limit, 0).Err()
	return err
}

func (t threshold) Remove() error {
	client, err := getConnection()
	if err != nil {
		fmt.Println("error: failed connecting to redis:", err)
		return err
	}
	err = client.Del(t.Key()).Err()
	return err
}

func (t threshold) Exceeded(rates []rate) bool {
	// we need at least 5 samples to determine a trend
	if len(rates) < 5 {
		return false
	}
	for i := len(rates) - 1; i > len(rates)-6; i-- {
		if rates[i].Value >= t.Limit {
			continue
		}
		return false
	}
	return true
}

func (t threshold) Alert() error {
	// TODO: implement sending alert through pushbullet API
	return nil
}

func newThreshold(email string, limit float64, maturity string) threshold {
	return threshold{Email: strings.Trim(email, "<>"), Limit: limit, Maturity: maturity}
}

func newThresholdFromKeyVal(key string, value float64) (threshold, error) {
	key = strings.TrimLeft(key, "gorates_")
	parts := strings.Split(key, ";")
	if len(parts) != 2 {
		return threshold{}, fmt.Errorf("email and maturity not found in key")
	}

	return threshold{Email: parts[0], Maturity: parts[1], Limit: value}, nil
}
