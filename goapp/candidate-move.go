package main

import (
	"sync"
	"time"
)

type CandidateMove struct {
	Move     string
	IssuedAt time.Time
	mu       *sync.Mutex
}

func (c *CandidateMove) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.Move = ""
	c.IssuedAt = time.Now()
}
