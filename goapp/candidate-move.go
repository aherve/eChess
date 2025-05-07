package main

import (
	"sync"
	"time"

	"github.com/aherve/eChess/goapp/lichess"
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

func (cm *CandidateMove) Move() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.move
}

func (cm *CandidateMove) IssuedAt() time.Time {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.issuedAt
}

func (cm *CandidateMove) Reset() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.move = ""
	cm.issuedAt = time.Now()
}

/*
* Will schedule a move and play it later, provided a new move hasn't been planned in between.
This method can be called on empty string to cancel a previously planned move
*/
func (cm *CandidateMove) PlayWithDelay(gameID, move string) {
	cm.recursivePlayWithDelay(gameID, move, true)
}

func (cm *CandidateMove) recursivePlayWithDelay(gameID, move string, shouldSchedule bool) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	existing := cm.move

	// this is a new move, and we should schedule it
	if move != existing && shouldSchedule {
		cm.move = move
		cm.issuedAt = time.Now()

		// schedule a new call (only if move isn't empty)
		if move != "" {
			go func(g, m string) {
				time.Sleep(PlayDelay + time.Millisecond)
				cm.recursivePlayWithDelay(g, m, false)
			}(gameID, move)
		}
		return
	}

	// This is not what was planned, and we should not reschedule => cancel
	if move != existing && !shouldSchedule {
		return
	}

	// move == existing

	// is it too soon ?
	if time.Since(cm.issuedAt) < PlayDelay {
		// same move is already scheduled for later. Don't pile up
		return
	}

	// empty move: we never send that to the server
	if move == "" {
		return
	}

	// move is non-empty, and it's time => play it!
	lichess.PlayMove(gameID, move)
	// reset our state
	cm.move = ""
	cm.issuedAt = time.Now()

}
