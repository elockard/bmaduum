package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewState(t *testing.T) {
	s := NewState()
	assert.NotNil(t, s)
	assert.False(t, s.IsDetected())
	assert.True(t, s.GetResetTime().IsZero())
	assert.Empty(t, s.GetLastError())
}

func TestState_MarkDetected(t *testing.T) {
	s := NewState()

	resetTime := time.Now().Add(5 * time.Minute)
	errorMsg := "rate limit reached"

	s.MarkDetected(resetTime, errorMsg)

	assert.True(t, s.IsDetected())
	assert.Equal(t, resetTime, s.GetResetTime())
	assert.Equal(t, errorMsg, s.GetLastError())
}

func TestState_IsReset(t *testing.T) {
	t.Run("not detected", func(t *testing.T) {
		s := NewState()
		assert.False(t, s.IsReset())
	})

	t.Run("reset time in future", func(t *testing.T) {
		s := NewState()
		future := time.Now().Add(5 * time.Minute)
		s.MarkDetected(future, "rate limit")
		assert.False(t, s.IsReset())
	})

	t.Run("reset time in past", func(t *testing.T) {
		s := NewState()
		past := time.Now().Add(-5 * time.Minute)
		s.MarkDetected(past, "rate limit")
		assert.True(t, s.IsReset())
	})
}

func TestState_WaitTime(t *testing.T) {
	t.Run("not detected", func(t *testing.T) {
		s := NewState()
		assert.Equal(t, time.Duration(0), s.WaitTime())
	})

	t.Run("reset time in future", func(t *testing.T) {
		s := NewState()
		future := time.Now().Add(5 * time.Minute)
		s.MarkDetected(future, "rate limit")
		wait := s.WaitTime()
		assert.True(t, wait > 0)
		// Should be approximately 5 minutes plus buffer
		assert.True(t, wait > 4*time.Minute)
	})

	t.Run("reset time in past", func(t *testing.T) {
		s := NewState()
		past := time.Now().Add(-5 * time.Minute)
		s.MarkDetected(past, "rate limit")
		assert.Equal(t, time.Duration(0), s.WaitTime())
	})
}

func TestState_Clear(t *testing.T) {
	s := NewState()

	// Set some state
	future := time.Now().Add(5 * time.Minute)
	s.MarkDetected(future, "rate limit")
	assert.True(t, s.IsDetected())

	// Clear it
	s.Clear()

	// Verify cleared
	assert.False(t, s.IsDetected())
	assert.True(t, s.GetResetTime().IsZero())
	assert.Empty(t, s.GetLastError())
}

func TestState_ConcurrentAccess(t *testing.T) {
	// This test ensures the mutex works correctly
	s := NewState()

	done := make(chan bool)

	// Writer goroutine
	go func() {
		for i := 0; i < 100; i++ {
			s.MarkDetected(time.Now(), "error")
			s.Clear()
		}
		done <- true
	}()

	// Reader goroutine
	go func() {
		for i := 0; i < 100; i++ {
			_ = s.IsDetected()
			_ = s.GetResetTime()
			_ = s.GetLastError()
			_ = s.IsReset()
			_ = s.WaitTime()
		}
		done <- true
	}()

	// Wait for both to complete
	<-done
	<-done

	// If we get here without deadlock or panic, the test passes
	assert.True(t, true)
}
