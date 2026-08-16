package rtorrent

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	invalidStringCase = "invalid string"
	invalidTypeCase   = "invalid type"
	stringCase        = "string"
)

func TestBoolFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
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
			t.Parallel()

			result, err := boolFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}

func TestIntFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected int
		errs     []error
	}{
		{"int", 1, 1, nil},
		{"int64", int64(1), 1, nil},
		{"float64", 1.0, 1, nil},
		{stringCase, "1", 1, nil},
		// A bad parse reports both our own sentinel and the underlying strconv error, courtesy of multi-%w
		{invalidStringCase, "invalid", 0, []error{ErrBadData, strconv.ErrSyntax}},
		{invalidTypeCase, []int{1}, 0, []error{ErrBadData}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := intFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.errs == nil {
				require.NoError(t, err)
				return
			}
			for _, want := range tt.errs {
				assert.ErrorIs(t, err, want)
			}
		})
	}
}

func TestTimeFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected time.Time
		errs     []error
	}{
		{"int", 1, time.Unix(1, 0), nil},
		{"int64", int64(1), time.Unix(1, 0), nil},
		{"float64", 1.0, time.Unix(1, 0), nil},
		{stringCase, "1", time.Unix(1, 0), nil},
		{invalidStringCase, "foo bar baz", time.Time{}, []error{ErrBadData, strconv.ErrSyntax}},
		{invalidTypeCase, []int{1}, time.Time{}, []error{ErrBadData}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := timeFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			if tt.errs == nil {
				require.NoError(t, err)
				return
			}
			for _, want := range tt.errs {
				assert.ErrorIs(t, err, want)
			}
		})
	}
}

func TestStringFromAny(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    any
		expected string
		err      error
	}{
		{stringCase, "test", "test", nil},
		{invalidTypeCase, 1, "", ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := stringFromAny(tt.input)
			assert.Equal(t, tt.expected, result)
			assert.ErrorIs(t, err, tt.err)
		})
	}
}
