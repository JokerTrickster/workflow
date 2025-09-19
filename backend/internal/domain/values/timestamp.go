package values

import (
	"errors"
	"time"
)

// Timestamp represents a domain-specific time value
type Timestamp struct {
	value time.Time
}

// NewTimestamp creates a new Timestamp value object
func NewTimestamp(value time.Time) (*Timestamp, error) {
	if err := validateTimestamp(value); err != nil {
		return nil, err
	}
	
	return &Timestamp{value: value}, nil
}

// Now creates a new Timestamp with the current time
func Now() *Timestamp {
	return &Timestamp{value: time.Now().UTC()}
}

// Value returns the time.Time value of the Timestamp
func (t Timestamp) Value() time.Time {
	return t.value
}

// UTC returns the timestamp in UTC
func (t Timestamp) UTC() time.Time {
	return t.value.UTC()
}

// Unix returns the Unix timestamp
func (t Timestamp) Unix() int64 {
	return t.value.Unix()
}

// String implements the Stringer interface
func (t Timestamp) String() string {
	return t.value.Format(time.RFC3339)
}

// Before checks if this timestamp is before another
func (t Timestamp) Before(other Timestamp) bool {
	return t.value.Before(other.value)
}

// After checks if this timestamp is after another
func (t Timestamp) After(other Timestamp) bool {
	return t.value.After(other.value)
}

// Equals checks if two timestamps are equal
func (t Timestamp) Equals(other Timestamp) bool {
	return t.value.Equal(other.value)
}

// Add adds a duration to the timestamp
func (t Timestamp) Add(duration time.Duration) Timestamp {
	return Timestamp{value: t.value.Add(duration)}
}

// Sub subtracts another timestamp and returns the duration
func (t Timestamp) Sub(other Timestamp) time.Duration {
	return t.value.Sub(other.value)
}

// validateTimestamp validates the timestamp value
func validateTimestamp(value time.Time) error {
	if value.IsZero() {
		return errors.New("timestamp cannot be zero")
	}
	
	// Check if timestamp is too far in the future (1 year from now)
	if value.After(time.Now().Add(365 * 24 * time.Hour)) {
		return errors.New("timestamp cannot be more than 1 year in the future")
	}
	
	// Check if timestamp is too far in the past (10 years ago)
	if value.Before(time.Now().Add(-10 * 365 * 24 * time.Hour)) {
		return errors.New("timestamp cannot be more than 10 years in the past")
	}
	
	return nil
}