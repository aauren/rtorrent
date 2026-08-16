package rtorrent

import (
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	invalidStringCase = "invalid string"
	invalidTypeCase   = "invalid type"
	stringCase        = "string"
)

func TestBoolFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected bool
		err      error
	}{
		{"bool true", true, true, nil},
		{"bool false", false, false, nil},
		{"int 1", 1, true, nil},
		{"int 0", 0, false, nil},
		{"int64 1", int64(1), true, nil},
		{"int64 0", int64(0), false, nil},
		{"float64 1.0", 1.0, true, nil},
		{"float64 0.0", 0.0, false, nil},
		{"string true", "true", true, nil},
		{"string false", "false", false, nil},
		{invalidStringCase, "invalid", false, ErrBadData},
		{invalidTypeCase, []int{1}, false, ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := boolFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.err, err)
		})
	}
}

func TestIntFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected int
		err      error
	}{
		{"int", 1, 1, nil},
		{"int64", int64(1), 1, nil},
		{"float64", 1.0, 1, nil},
		{stringCase, "1", 1, nil},
		{invalidStringCase, "invalid", 0, strconv.ErrSyntax},
		{invalidTypeCase, []int{1}, 0, ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := intFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.True(t, errors.Is(err, tt.err), "expected %v, got %v", tt.err, err)
		})
	}
}

func TestTimeFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected time.Time
		err      error
	}{
		{"int", 1, time.Unix(1, 0), nil},
		{"int64", int64(1), time.Unix(1, 0), nil},
		{"float64", 1.0, time.Unix(1, 0), nil},
		{stringCase, "1", time.Unix(1, 0), nil},
		{invalidStringCase, "foo bar baz", time.Time{}, ErrBadData},
		{invalidTypeCase, []int{1}, time.Time{}, ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := timeFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.True(t, errors.Is(err, tt.err), "expected %v, got %v", tt.err, err)
		})
	}
}

func TestStringFromAny(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
		err      error
	}{
		{stringCase, "test", "test", nil},
		{invalidTypeCase, 1, "", ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := stringFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.Equal(t, tt.err, err)
		})
	}
}
