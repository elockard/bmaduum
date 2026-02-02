// Package ratelimit provides rate limit detection for Claude CLI.
//
// The detector parses stderr output from Claude CLI to identify rate limit
// errors and extract the reset time. This enables automatic retry with
// intelligent wait times.
package ratelimit

import (
	"regexp"
	"strings"
	"time"
)

// ErrorInfo contains parsed information from a rate limit error.
type ErrorInfo struct {
	// IsRateLimit is true if this is a rate limit error.
	IsRateLimit bool

	// ResetTime is the time when the rate limit will reset.
	// Only valid if IsRateLimit is true.
	ResetTime time.Time

	// RawMessage is the original error message.
	RawMessage string
}

// Detector detects rate limit errors from Claude CLI stderr output.
//
// The detector uses regex patterns to identify rate limit error messages
// and extract the reset time from them.
type Detector struct {
	// rateLimitPattern matches rate limit error messages.
	// Example: "Claude usage limit reached. Your limit will reset at 1pm (Etc/GMT+5)"
	rateLimitPattern *regexp.Regexp

	// resetTimePattern extracts the reset time from rate limit messages.
	resetTimePattern *regexp.Regexp
}

// NewDetector creates a new rate limit detector.
func NewDetector() *Detector {
	return &Detector{
		// Match rate limit messages like:
		// "Claude usage limit reached. Your limit will reset at 1pm (Etc/GMT+5)"
		rateLimitPattern: regexp.MustCompile(`(?i)usage limit reached|rate limit|quota exceeded`),

		// Extract time from messages like:
		// "Your limit will reset at 1pm (Etc/GMT+5)" - captures just the time part
		resetTimePattern: regexp.MustCompile(`reset at ([^([]+)`),
	}
}

// CheckLine checks a single line of stderr output for rate limit errors.
//
// Returns an ErrorInfo with IsRateLimit=true if a rate limit error is detected.
// The ResetTime field will be populated if the time can be parsed from the message.
func (d *Detector) CheckLine(line string) ErrorInfo {
	if !d.rateLimitPattern.MatchString(line) {
		return ErrorInfo{IsRateLimit: false}
	}

	info := ErrorInfo{
		IsRateLimit: true,
		RawMessage:  line,
	}

	// Try to extract reset time
	if matches := d.resetTimePattern.FindStringSubmatch(line); len(matches) > 1 {
		// Attempt to parse various time formats
		resetTime := d.parseResetTime(strings.TrimSpace(matches[1]))
		if !resetTime.IsZero() {
			info.ResetTime = resetTime
		}
	}

	return info
}

// parseResetTime attempts to parse the reset time from a time string.
//
// Returns the parsed time or zero time if parsing fails.
func (d *Detector) parseResetTime(timeStr string) time.Time {
	// List of formats to try
	formats := []string{
		"3:04pm",
		"3:04 pm",
		"3:04PM",
		"3:04 PM",
		"15:04",
		"3pm",
		"3 pm",
		"3PM",
		"3 PM",
	}

	// Try parsing as today with the given time
	now := time.Now()
	for _, format := range formats {
		t, err := time.Parse(format, timeStr)
		if err == nil {
			// Combine with today's date
			return time.Date(
				now.Year(), now.Month(), now.Day(),
				t.Hour(), t.Minute(), 0, 0,
				now.Location(),
			)
		}
	}

	return time.Time{}
}

// WaitTime calculates how long to wait before retrying.
//
// If the ErrorInfo has a valid ResetTime, returns the duration until that time.
// Otherwise, returns a default wait time.
func (d *Detector) WaitTime(info ErrorInfo) time.Duration {
	if !info.ResetTime.IsZero() {
		wait := time.Until(info.ResetTime)
		if wait > 0 {
			return wait + 30*time.Second // Add buffer
		}
	}

	// Default wait time if we couldn't parse the reset time
	return 5 * time.Minute
}
