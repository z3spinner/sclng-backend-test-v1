package fetcherMock

import (
	"container/list"
	"sync"
	"time"

	"github.com/Scalingo/sclng-backend-test-v1/worker/interfaces/fetcher"
)

type mockRateLimiter struct {
	mutex      sync.Mutex
	limit      int
	resetTimes *list.List
	timeout    time.Duration
}

func newRateLimiter(limit int, timeout time.Duration) *mockRateLimiter {
	return &mockRateLimiter{
		mutex:      sync.Mutex{},
		limit:      limit,
		resetTimes: list.New(),
		timeout:    timeout,
	}
}

// purgeResetTimes removes all reset times that are in the past by removing them from the front of the list
func (r *mockRateLimiter) purgeResetTimes() {

	// remove all reset times that are in the past
	for e := r.resetTimes.Front(); e != nil; {
		next := e.Next()

		// The first item in the list is later than now, so we can stop
		if !e.Value.(time.Time).Before(time.Now()) {
			break
		}

		// Remove the item from the list, because it is in the past
		r.resetTimes.Remove(e)
		e = next
	}

}

// remaining returns the number of remaining calls available
func (r *mockRateLimiter) remaining() int {
	return r.limit - r.resetTimes.Len()
}

// resetTime returns the time when the rate limit will have some calls remaining
func (r *mockRateLimiter) resetTime() time.Time {
	var resetTime time.Time
	if r.resetTimes.Len() > 0 {
		resetTime = r.resetTimes.Front().Value.(time.Time)
	}
	return resetTime
}

// get returns the remaining rate limit and the resetTimes time
func (r *mockRateLimiter) get() (int, time.Time) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	return r.remaining(), r.resetTime()
}

// checkMockRateLimiting checks the mock rate limiter
// The mock rate limiter is also updated with a new reset time as if a call had been made to a real API
// It returns an error if the rate limit has been exceeded
func (r *mockRateLimiter) checkMockRateLimiting() error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	// Clear out old requests
	r.purgeResetTimes()

	// If we have no calls remaining, then return a rate limited error
	//fmt.Println("r.remaining(): ", r.remaining(), r.resetTimes.Len())
	if r.remaining() < 1 {
		return fetcher.ErrRateLimited
	}

	// Add a new reset time to the end of the list
	r.resetTimes.PushBack(time.Now().Add(r.timeout))

	return nil
}
