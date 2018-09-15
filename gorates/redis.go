package main

import (
	"fmt"
	"sync"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	redisAddr = "localhost:6379"
	retryConn = 5
)

var (
	redisCli   *redis.Client
	redisMutex = &sync.Mutex{}
)

func getConnection() (*redis.Client, error) {
	redisMutex.Lock()
	defer redisMutex.Unlock()

	var err error
	for i := 0; i < retryConn; i++ {
		redisCli, err = checkConnection(redisCli)
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100) // sleep 0.1 seconds between retries
	}
	return redisCli, err
}

func checkConnection(cli *redis.Client) (*redis.Client, error) {
	if cli == nil || cli.Ping().Err() != nil {
		// clean up the old connection
		if cli != nil {
			cli.Close()
		}

		// open a new connection
		client := redis.NewClient(&redis.Options{Addr: redisAddr, DB: 0})

		// check that the new connection works
		err := client.Ping().Err()
		if err != nil {
			client.Close()
			return nil, err
		}

		return client, nil
	}

	return cli, nil
}

func addThreshold(th threshold) error {
	client, err := getConnection()
	if err != nil {
		fmt.Println("error: failed connecting to redis:", err)
		return err
	}
	return cacheValue(client, th.Key(), th.Limit)
}

func removeThreshold(th threshold) error {
	client, err := getConnection()
	if err != nil {
		fmt.Println("error: failed connecting to redis:", err)
		return err
	}
	err = client.Del(th.Key()).Err()
	return err
}

func loadThresholds(email string) []threshold {
	client, err := getConnection()
	if err != nil {
		fmt.Println("error: failed connecting to redis:", err)
		return []threshold{}
	}

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

func cacheValue(client *redis.Client, key string, value float64) error {
	err := client.Set(key, value, 0).Err()
	return err
}
