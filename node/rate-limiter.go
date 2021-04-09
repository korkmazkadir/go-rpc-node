package node

import "log"

type job struct {
	op       func()
	cost     int
	waitChan chan struct{}
}

// RateLimiter implements a rate limiter
type RateLimiter struct {
	maxTokens       int
	availableTokens int
	sleepCount      int

	returnedTokens chan int
	operationChan  chan job
}

// NewRateLimiter creates an instance of rate limiter and returns a pointer to it
func NewRateLimiter(maxTokens int) *RateLimiter {
	limiter := &RateLimiter{maxTokens: maxTokens, availableTokens: maxTokens}
	limiter.returnedTokens = make(chan int, 100)
	limiter.operationChan = make(chan job, 100)
	go limiter.mainLoop()
	return limiter
}

// ApplyRateLimit applies ratelimite to exceute the provided function
func (r *RateLimiter) ApplyRateLimit(operation func(), cost int) {
	waitChan := make(chan struct{}, 1)
	r.operationChan <- job{op: operation, cost: cost, waitChan: waitChan}
	<-waitChan
}

func (r *RateLimiter) accuireReturnedTokens() {
	for {
		select {
		case t := <-r.returnedTokens:
			r.availableTokens += t
			if r.availableTokens > 0 {
				return
			}
		}
	}
}

func (r *RateLimiter) mainLoop() {

	for {
		select {
		case operation := <-r.operationChan:

			log.Println(r.availableTokens)

			if r.availableTokens < 0 {
				// waits for enough token
				r.accuireReturnedTokens()
			}

			r.availableTokens = r.availableTokens - operation.cost

			// apply operation
			go func(operation job) {
				operation.op()
				// return tokens
				r.returnedTokens <- operation.cost
				operation.waitChan <- struct{}{}
			}(operation)

		}
	}

}
