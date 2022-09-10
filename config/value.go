package config

import (
	"encoding/base64"
	"strconv"
	"time"
)

// bytesValue implements encoding.TextMarshaler and encoding.TextUnmarshaler. Usually used in flag.TextVar() function.
type bytesValue []byte

func (v *bytesValue) MarshalText() ([]byte, error) {
	text := make([]byte, base64.StdEncoding.EncodedLen(len(*v)))
	base64.StdEncoding.Encode(text, *v)
	return text, nil
}

func (v *bytesValue) UnmarshalText(text []byte) error {
	*v = make([]byte, base64.StdEncoding.DecodedLen(len(text)))
	n, err := base64.StdEncoding.Decode(*v, text)
	if err != nil {
		return err
	}
	*v = (*v)[:n]
	return nil
}

// parseDefaultBool creates named return values and returns default bool value by naked return statement if input string is empty.
func parseDefaultBool(s string) (b bool, err error) {
	if s != "" {
		return strconv.ParseBool(s)
	}
	return
}

// parseDefaultInt creates named return values and returns default int value by naked return statement if input string is empty.
func parseDefaultInt(s string) (i int, err error) {
	if s != "" {
		v, err := strconv.ParseInt(s, 0, strconv.IntSize)
		return int(v), err
	}
	return
}

// parseDefaultInt64 creates named return values and returns default int64 value by naked return statement if input string is empty.
func parseDefaultInt64(s string) (i int64, err error) {
	if s != "" {
		return strconv.ParseInt(s, 0, 64)
	}
	return
}

// parseDefaultUint creates named return values and returns default uint value by naked return statement if input string is empty.
func parseDefaultUint(s string) (u uint, err error) {
	if s != "" {
		v, err := strconv.ParseUint(s, 0, strconv.IntSize)
		return uint(v), err
	}
	return
}

// parseDefaultUint64 creates named return values and returns default uint64 value by naked return statement if input string is empty.
func parseDefaultUint64(s string) (u uint64, err error) {
	if s != "" {
		return strconv.ParseUint(s, 0, 64)
	}
	return
}

// parseDefaultFloat64 creates named return values and returns default float64 value by naked return statement if input string is empty.
func parseDefaultFloat64(s string) (f float64, err error) {
	if s != "" {
		return strconv.ParseFloat(s, 64)
	}
	return
}

// parseDefaultDuration creates named return values and returns default time.Duration value by naked return statement if input string is empty.
func parseDefaultDuration(s string) (d time.Duration, err error) {
	if s != "" {
		return time.ParseDuration(s)
	}
	return
}

// parseDefaultBytesValue returns empty bytes slice as default bytesValue value if input string is empty.
func parseDefaultBytesValue(s string) (b *bytesValue, err error) {
	b = &bytesValue{}
	if s != "" {
		err = b.UnmarshalText([]byte(s))
		return b, err
	}
	return
}
