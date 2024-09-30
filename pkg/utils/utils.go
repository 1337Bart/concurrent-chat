package utils

import (
	"errors"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// GenerateRandomID generates a random string of specified length
func GenerateRandomID(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// TimeoutAfter runs a function with a timeout
func TimeoutAfter(d time.Duration, f func() error) error {
	errChan := make(chan error, 1)
	go func() {
		errChan <- f()
	}()

	select {
	case err := <-errChan:
		return err
	case <-time.After(d):
		return ErrTimeout
	}
}

var ErrTimeout = errors.New("operation timed out")
