package entities

import (
	"sync"
	"time"
)

// RateLimits is a struct to hold rate limit data
type RateLimits struct {
	remaining int
	reset     time.Time
	mutex     *sync.Mutex
}

// NewRateLimits creates a new RateLimits struct
func NewRateLimits() RateLimits {
	return RateLimits{
		mutex: &sync.Mutex{},
	}
}

// SetRateLimits sets the rate limits
func (c *RateLimits) SetRateLimits(remaining int, reset time.Time) {
	// lock the mutex
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.remaining = remaining
	c.reset = reset

	return
}

func (c *RateLimits) GetRemainingCount() int {
	return c.remaining
}

func (c *RateLimits) GetResetTime() time.Time {
	return c.reset
}
