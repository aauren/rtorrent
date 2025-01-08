package rtorrent

import (
	"strconv"
	"time"
)

func boolFromAny(data interface{}) (bool, error) {
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
			return false, ErrBadData
		}
		return b, nil
	default:
		return false, ErrBadData
	}
}

func intFromAny(data interface{}) (int, error) {
	switch v := data.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	case string:
		return strconv.Atoi(v)
	default:
		return 0, ErrBadData
	}
}

func timeFromAny(data interface{}) (time.Time, error) {
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
			return time.Time{}, ErrBadData
		}
		return time.Unix(timeInt, 0), nil
	default:
		return time.Time{}, ErrBadData
	}
}

func stringFromAny(data interface{}) (string, error) {
	switch v := data.(type) {
	case string:
		return v, nil
	default:
		return "", ErrBadData
	}
}
