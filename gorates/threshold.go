package main

import (
	"fmt"
	"strings"

	"github.com/gregdel/pushover"
	redis "gopkg.in/redis.v5"
)

type threshold struct {
	UserToken string  `json:"user_token"`
	Limit     float64 `json:"limit"`
	Maturity  string  `json:"maturity"`
	Triggerd  bool    `json:"-"`
	Date      string  `json:"-"`
}

func (t threshold) Key() string {
	return fmt.Sprintf("gorates_%s;%s", t.UserToken, t.Maturity)
}

func (t threshold) Add(client *redis.Client) error {
	err := client.Set(t.Key(), t.Limit, 0).Err()
	return err
}

func (t threshold) Remove(client *redis.Client) error {
	err := client.Del(t.Key()).Err()
	return err
}

func (t *threshold) Exceeded(rates []rate) bool {
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

	t.Date = rates[len(rates)-1].Date.Format("2006-01-02")
	return true
}

func (t threshold) Alert(pushoverAppToken string) error {
	if pushoverAppToken == "" {
		return fmt.Errorf("pushover app token not configured")
	}

	app := pushover.New(pushoverAppToken)
	recipient := pushover.NewRecipient(t.UserToken)
	title := "Automatic Euribor alert"
	body := fmt.Sprintf("Your defined limit '%.3f' for Euribor rate %s has been exceeded at %s", t.Limit, t.Maturity, t.Date)
	message := pushover.NewMessageWithTitle(body, title)
	_, err := app.SendMessage(message, recipient)
	if err != nil {
		return err
	}

	return nil
}

func newThreshold(userToken string, limit float64, maturity string) threshold {
	return threshold{UserToken: userToken, Limit: limit, Maturity: maturity}
}

func newThresholdFromKeyVal(key string, value float64) (threshold, error) {
	key = strings.TrimLeft(key, "gorates_")
	fmt.Println(key)
	parts := strings.Split(key, ";")
	if len(parts) != 2 {
		return threshold{}, fmt.Errorf("user_token and maturity not found in key")
	}

	return threshold{UserToken: parts[0], Maturity: parts[1], Limit: value}, nil
}

func loadThresholds(client *redis.Client, userToken string) []threshold {
	keys, err := client.Keys("gorates_*").Result()
	if err != nil {
		fmt.Println("error: failed retrieving keys from redis:", err)
		return []threshold{}
	}

	thresholds := []threshold{}
	for _, key := range keys {
		value, err := lookupValue(client, key)
		if err != nil {
			fmt.Printf("error: failed looking up value for key '%s': %v\n", key, err)
			continue
		}
		threshold, err := newThresholdFromKeyVal(key, value)
		if err != nil {
			fmt.Printf("error: creating threshold from key '%s' and value '%v': %v\n", key, value, err)
			continue
		}
		if userToken != "" {
			if threshold.UserToken == userToken {
				thresholds = append(thresholds, threshold)
			}
		} else {
			thresholds = append(thresholds, threshold)
		}
	}
	return thresholds
}

func lookupValue(client *redis.Client, key string) (float64, error) {
	val, err := client.Get(key).Float64()
	if err != nil {
		return 0, err
	}
	return val, nil
}
