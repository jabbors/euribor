package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	redis "gopkg.in/redis.v5"
)

const (
	pushbulletURL = "https://api.pushbullet.com/v2/pushes"
)

type threshold struct {
	Email    string  `json:"email"`
	Limit    float64 `json:"limit"`
	Maturity string  `json:"maturity"`
	Triggerd bool    `json:"-"`
	Date     string  `json:"-"`
}

func (t threshold) Key() string {
	return fmt.Sprintf("gorates_%s;%s", t.Email, t.Maturity)
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

func (t threshold) Alert(pushbulletToken string) error {
	if pushbulletToken == "" {
		return fmt.Errorf("pushbullet token not configured")
	}

	data := struct {
		Type  string `json:"type"`
		Title string `json:"title"`
		Body  string `json:"body"`
		Email string `json:"email"`
	}{
		"note",
		fmt.Sprint("Automatic Euribor alert"),
		fmt.Sprintf("Your defined limit '%.3f' for Euribor rate %s has been exceeded at %s", t.Limit, t.Maturity, t.Date),
		t.Email,
	}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", pushbulletURL, bytes.NewBuffer(jsonStr))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", pushbulletToken))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("returned code %d, expected %d", resp.StatusCode, http.StatusOK)
	}

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

func loadThresholds(client *redis.Client, email string) []threshold {
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
		if email != "" {
			if threshold.Email == email {
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
