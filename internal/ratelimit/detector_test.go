package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewDetector(t *testing.T) {
	d := NewDetector()
	assert.NotNil(t, d)
	assert.NotNil(t, d.rateLimitPattern)
	assert.NotNil(t, d.resetTimePattern)
}

func TestDetector_CheckLine(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name          string
		line          string
		wantIsLimit   bool
		wantResetTime bool // whether we expect a reset time to be parsed
	}{
		{
			name:          "rate limit with time",
			line:          "Claude usage limit reached. Your limit will reset at 1pm (Etc/GMT+5)",
			wantIsLimit:   true,
			wantResetTime: true,
		},
		{
			name:          "rate limit lowercase",
			line:          "usage limit reached, please try again later",
			wantIsLimit:   true,
			wantResetTime: false,
		},
		{
			name:          "quota exceeded",
			line:          "Quota exceeded for this operation",
			wantIsLimit:   true,
			wantResetTime: false,
		},
		{
			name:          "normal error",
			line:          "Error: connection failed",
			wantIsLimit:   false,
			wantResetTime: false,
		},
		{
			name:          "empty line",
			line:          "",
			wantIsLimit:   false,
			wantResetTime: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := d.CheckLine(tt.line)
			assert.Equal(t, tt.wantIsLimit, info.IsRateLimit)
			if tt.wantIsLimit {
				assert.Equal(t, tt.line, info.RawMessage)
			}
			if tt.wantResetTime {
				assert.False(t, info.ResetTime.IsZero(), "expected reset time to be parsed")
			}
		})
	}
}

func TestDetector_parseResetTime(t *testing.T) {
	d := NewDetector()

	tests := []struct {
		name      string
		timeStr   string
		wantValid bool
	}{
		{"3:04pm format", "3:04pm", true},
		{"3:04 pm format", "3:04 pm", true},
		{"3pm format", "3pm", true},
		{"15:04 format", "15:04", true},
		{"invalid format", "not-a-time", false},
		{"empty string", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := d.parseResetTime(tt.timeStr)
			if tt.wantValid {
				assert.False(t, result.IsZero(), "expected time to be parsed")
			} else {
				assert.True(t, result.IsZero(), "expected time to be zero")
			}
		})
	}
}

func TestDetector_WaitTime(t *testing.T) {
	d := NewDetector()

	t.Run("with reset time in future", func(t *testing.T) {
		future := time.Now().Add(5 * time.Minute)
		info := ErrorInfo{
			IsRateLimit: true,
			ResetTime:   future,
		}
		wait := d.WaitTime(info)
		assert.True(t, wait > 0, "expected positive wait time")
		assert.True(t, wait > 30*time.Second, "expected wait time to include buffer")
	})

	t.Run("with reset time in past", func(t *testing.T) {
		past := time.Now().Add(-5 * time.Minute)
		info := ErrorInfo{
			IsRateLimit: true,
			ResetTime:   past,
		}
		wait := d.WaitTime(info)
		assert.Equal(t, 5*time.Minute, wait, "expected default wait time when reset is in past")
	})

	t.Run("without reset time", func(t *testing.T) {
		info := ErrorInfo{
			IsRateLimit: true,
			ResetTime:   time.Time{},
		}
		wait := d.WaitTime(info)
		assert.Equal(t, 5*time.Minute, wait, "expected default wait time")
	})
}

func TestErrorInfo_Fields(t *testing.T) {
	info := ErrorInfo{
		IsRateLimit: true,
		ResetTime:   time.Now(),
		RawMessage:  "test message",
	}

	assert.True(t, info.IsRateLimit)
	assert.False(t, info.ResetTime.IsZero())
	assert.Equal(t, "test message", info.RawMessage)
}
