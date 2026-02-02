package ratelimit

import (
	"sync"
	"time"
)

// State provides thread-safe rate limit state management.
//
// State tracks whether a rate limit has been detected and when it's expected
// to reset. It is safe for concurrent use.
type State struct {
	mu sync.RWMutex

	// detected indicates whether a rate limit error has been detected.
	detected bool

	// resetTime is when the rate limit is expected to reset.
	resetTime time.Time

	// lastError stores the last rate limit error message.
	lastError string
}

// NewState creates a new rate limit state manager.
func NewState() *State {
	return &State{}
}

// MarkDetected marks that a rate limit error has been detected.
//
// The resetTime parameter indicates when the rate limit is expected to reset.
// The errorMsg is stored for diagnostic purposes.
func (s *State) MarkDetected(resetTime time.Time, errorMsg string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.detected = true
	s.resetTime = resetTime
	s.lastError = errorMsg
}

// IsDetected returns true if a rate limit error has been detected.
func (s *State) IsDetected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.detected
}

// GetResetTime returns the expected rate limit reset time.
//
// Returns zero time if no rate limit has been detected.
func (s *State) GetResetTime() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.resetTime
}

// GetLastError returns the last rate limit error message.
func (s *State) GetLastError() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.lastError
}

// IsReset returns true if the rate limit has reset (current time is past reset time).
//
// Returns false if no rate limit has been detected.
func (s *State) IsReset() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.detected {
		return false
	}

	return time.Now().After(s.resetTime)
}

// Clear clears the rate limit detection state.
func (s *State) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.detected = false
	s.resetTime = time.Time{}
	s.lastError = ""
}

// WaitTime returns how long to wait before retrying.
//
// Returns 0 if the rate limit has already reset or no rate limit was detected.
// Returns the duration until reset plus a buffer if the rate limit is still active.
func (s *State) WaitTime() time.Duration {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.detected {
		return 0
	}

	wait := time.Until(s.resetTime) + 30*time.Second
	if wait < 0 {
		return 0
	}

	return wait
}
