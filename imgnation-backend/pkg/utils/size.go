package utils

import (
	"strconv"
	"strings"
)

const (
	UNIT_GB = "GB"
	UNIT_MB = "MB"
	UNIT_KB = "KB"
	UNIT_B  = "B"

	SIZE_GB int64 = 1 << 30
	SIZE_MB int64 = 1 << 20
	SIZE_KB int64 = 1 << 30
	SIZE_B  int64 = 1
)

func SizeMultiplierFromStr(unit string) int64 {
	switch unit {
	case UNIT_GB:
		return SIZE_GB
	case UNIT_MB:
		return SIZE_MB
	case UNIT_KB:
		return SIZE_KB
	default:
		return SIZE_B
	}
}

// returns -1 if failed
func ParseStrSize(s string) int64 {
	s = strings.ToUpper(strings.TrimSpace(s))
	if len(s) == 0 {
		return 0
	}
	var unit string
	for _, u := range []string{"GB", "MB", "KB", "B"} {
		if strings.HasSuffix(s, u) {
			unit = u
			s = s[:len(u)-1]
			break
		}
	}
	multiplier := SizeMultiplierFromStr(unit)
	size, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	if size < 0 {
		return -1
	}
	return size * multiplier
}

func Default[T comparable](val T, defaultVal T) T {
	var null T
	if val == null {
		return defaultVal
	}
	return val
}

func Ternar[T comparable](exp bool, t T, f T) T {
	if exp {
		return t
	}
	return f
}
