package main

import (
	"sync"
	"time"
)

type CandidateMove struct {
	move     string
	issuedAt time.Time
	mu       sync.RWMutex
}

func NewCandidateMove() *CandidateMove {
	return &CandidateMove{
		move:     "",
		issuedAt: time.Now(),
	}
}

func (c *CandidateMove) Move() string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.move
}

func (c *CandidateMove) IssuedAt() time.Time {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.issuedAt
}

func (c *CandidateMove) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.move = ""
	c.issuedAt = time.Now()
}

func (c *CandidateMove) Set(move string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.move = move
	c.issuedAt = time.Now()
}
