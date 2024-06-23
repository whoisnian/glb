package config

import (
	"encoding/base64"
	"errors"
	"reflect"
	"strconv"
	"time"
)

// Value is the interface to the dynamic value stored in a flag.
type Value interface {
	Type() string
	String() string
	IsZero() bool
	Set(string) error
}

func newFlagValue(v reflect.Value, defValue string) (value Value, err error) {
	switch pv := v.Addr().Interface().(type) {
	case *bool:
		value = (*boolValue)(pv)
	case *int:
		value = (*intValue)(pv)
	case *int64:
		value = (*int64Value)(pv)
	case *uint:
		value = (*uintValue)(pv)
	case *uint64:
		value = (*uint64Value)(pv)
	case *string:
		value = (*stringValue)(pv)
	case *float64:
		value = (*float64Value)(pv)
	case *time.Duration:
		value = (*durationValue)(pv)
	case *[]byte:
		value = (*bytesValue)(pv)
	default:
		return nil, errors.New("config: unknown value type " + v.Type().Kind().String())
	}
	return value, value.Set(defValue)
}

// -- bool Value
type boolValue bool

func (b *boolValue) Type() string   { return "bool" }
func (b *boolValue) String() string { return strconv.FormatBool(bool(*b)) }
func (b *boolValue) IsZero() bool   { return !bool(*b) }
func (b *boolValue) Set(s string) (err error) {
	var v bool // default `false` if input is empty
	if s != "" {
		v, err = strconv.ParseBool(s)
	}
	*b = boolValue(v)
	return err
}

func (b *boolValue) IsBoolFlag() bool { return true }

type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- int Value
type intValue int

func (i *intValue) Type() string   { return "int" }
func (i *intValue) String() string { return strconv.Itoa(int(*i)) }
func (i *intValue) IsZero() bool   { return *i == 0 }
func (i *intValue) Set(s string) (err error) {
	var v int64 // default `0` if input is empty
	if s != "" {
		v, err = strconv.ParseInt(s, 0, strconv.IntSize)
	}
	*i = intValue(v)
	return err
}

// -- int64 Value
type int64Value int64

func (i *int64Value) Type() string   { return "int64" }
func (i *int64Value) String() string { return strconv.FormatInt(int64(*i), 10) }
func (i *int64Value) IsZero() bool   { return *i == 0 }
func (i *int64Value) Set(s string) (err error) {
	var v int64 // default `0` if input is empty
	if s != "" {
		v, err = strconv.ParseInt(s, 0, 64)
	}
	*i = int64Value(v)
	return err
}

// -- uint Value
type uintValue uint

func (i *uintValue) Type() string   { return "uint" }
func (i *uintValue) String() string { return strconv.FormatUint(uint64(*i), 10) }
func (i *uintValue) IsZero() bool   { return *i == 0 }
func (i *uintValue) Set(s string) (err error) {
	var v uint64 // default `0` if input is empty
	if s != "" {
		v, err = strconv.ParseUint(s, 0, strconv.IntSize)
	}
	*i = uintValue(v)
	return err
}

// -- uint64 Value
type uint64Value uint64

func (i *uint64Value) Type() string   { return "uint64" }
func (i *uint64Value) String() string { return strconv.FormatUint(uint64(*i), 10) }
func (i *uint64Value) IsZero() bool   { return *i == 0 }
func (i *uint64Value) Set(s string) (err error) {
	var v uint64 // default `0` if input is empty
	if s != "" {
		v, err = strconv.ParseUint(s, 0, 64)
	}
	*i = uint64Value(v)
	return err
}

// -- string Value
type stringValue string

func (s *stringValue) Type() string   { return "string" }
func (s *stringValue) String() string { return string(*s) }
func (s *stringValue) IsZero() bool   { return *s == "" }
func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

// -- float64 Value
type float64Value float64

func (f *float64Value) Type() string   { return "float64" }
func (f *float64Value) String() string { return strconv.FormatFloat(float64(*f), 'g', -1, 64) }
func (f *float64Value) IsZero() bool   { return *f == 0 }
func (f *float64Value) Set(s string) (err error) {
	var v float64 // default `0` if input is empty
	if s != "" {
		v, err = strconv.ParseFloat(s, 64)
	}
	*f = float64Value(v)
	return err
}

// -- time.Duration Value
type durationValue time.Duration

func (d *durationValue) Type() string   { return "duration" }
func (d *durationValue) String() string { return (*time.Duration)(d).String() }
func (d *durationValue) IsZero() bool   { return *d == 0 }
func (d *durationValue) Set(s string) (err error) {
	var v time.Duration // default `0` if input is empty
	if s != "" {
		v, err = time.ParseDuration(s)
	}
	*d = durationValue(v)
	return err
}

// -- bytes Value
type bytesValue []byte

func (b *bytesValue) Type() string   { return "bytes" }
func (b *bytesValue) String() string { return base64.StdEncoding.EncodeToString([]byte(*b)) }
func (b *bytesValue) IsZero() bool   { return *b == nil }
func (b *bytesValue) Set(s string) (err error) {
	var v []byte // default `nil` if input is empty
	if s != "" {
		v, err = base64.StdEncoding.DecodeString(s)
	}
	*b = bytesValue(v)
	return err
}
