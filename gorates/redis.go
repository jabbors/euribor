package main

import (
	"fmt"
	"sync"
	"time"

	redis "gopkg.in/redis.v5"
)

const (
	retryConn = 5
)

// redisConnector represents a connection ta a redis DB
type redisConnector struct {
	host   string
	port   int
	client *redis.Client
	mutex  *sync.Mutex
}

// NewRedisConnector returns an initialzed connector for a redis DB
func NewRedisConnector(host string, port int) *redisConnector {
	rc := redisConnector{
		host:  host,
		port:  port,
		mutex: &sync.Mutex{},
	}

	return &rc
}

func (rc *redisConnector) Connect() (*redis.Client, error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	var err error
	for i := 0; i < retryConn; i++ {
		err = rc.check()
		if err == nil {
			break
		}
		time.Sleep(time.Millisecond * 100) // sleep 0.1 seconds between retries
	}
	return rc.client, err
}

func (rc *redisConnector) check() error {
	if rc.client == nil || rc.client.Ping().Err() != nil {
		// clean up the old connection
		if rc.client != nil {
			rc.client.Close()
		}

		// open a new connection
		redisAddr := fmt.Sprintf("%s:%d", rc.host, rc.port)
		client := redis.NewClient(&redis.Options{Addr: redisAddr, DB: 0})

		// check that the new connection works
		err := client.Ping().Err()
		if err != nil {
			client.Close()
			return err
		}
		rc.client = client

		return nil
	}

	return nil
}
