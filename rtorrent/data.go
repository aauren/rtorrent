package rtorrent

import (
	"fmt"
	"strconv"
	"time"
)

// Converters wrap failures in ErrBadData for a single errors.Is check, while multi-%w keeps the strconv error reachable

func boolFromAny(data any) (bool, error) {
	switch v := data.(type) {
	case bool:
		return v, nil
	case int:
		return v == 1, nil
	case int64:
		return v == 1, nil
	case float64:
		return v == 1.0, nil
	case string:
		b, err := strconv.ParseBool(v)
		if err != nil {
			return false, fmt.Errorf("%w: %w", ErrBadData, err)
		}
		return b, nil
	default:
		return false, fmt.Errorf("%w: cannot convert %T to bool", ErrBadData, data)
	}
}

func intFromAny(data any) (int, error) {
	switch v := data.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		i, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("%w: %w", ErrBadData, err)
		}
		return i, nil
	default:
		return 0, fmt.Errorf("%w: cannot convert %T to int", ErrBadData, data)
	}
}

func timeFromAny(data any) (time.Time, error) {
	switch v := data.(type) {
	case int:
		return time.Unix(int64(v), 0), nil
	case int64:
		return time.Unix(v, 0), nil
	case float64:
		return time.Unix(int64(v), 0), nil
	case string:
		timeInt, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("%w: %w", ErrBadData, err)
		}
		return time.Unix(timeInt, 0), nil
	default:
		return time.Time{}, fmt.Errorf("%w: cannot convert %T to time.Time", ErrBadData, data)
	}
}

func stringFromAny(data any) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	default:
		return "", fmt.Errorf("%w: cannot convert %T to string", ErrBadData, data)
	}
}
