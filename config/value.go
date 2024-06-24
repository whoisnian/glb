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
	IsZero(string) bool
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

func (v *boolValue) Type() string         { return "bool" }
func (v *boolValue) String() string       { return strconv.FormatBool(bool(*v)) }
func (v *boolValue) IsZero(s string) bool { return s == "false" }
func (v *boolValue) Set(s string) (err error) {
	var res bool // default `false` if input is empty
	if s != "" {
		res, err = strconv.ParseBool(s)
	}
	*v = boolValue(res)
	return err
}

func (v *boolValue) IsBoolFlag() bool { return true }

type boolFlag interface {
	Value
	IsBoolFlag() bool
}

// -- int Value
type intValue int

func (v *intValue) Type() string         { return "int" }
func (v *intValue) String() string       { return strconv.Itoa(int(*v)) }
func (v *intValue) IsZero(s string) bool { return s == "0" }
func (v *intValue) Set(s string) (err error) {
	var res int64 // default `0` if input is empty
	if s != "" {
		res, err = strconv.ParseInt(s, 0, strconv.IntSize)
	}
	*v = intValue(res)
	return err
}

// -- int64 Value
type int64Value int64

func (v *int64Value) Type() string         { return "int64" }
func (v *int64Value) String() string       { return strconv.FormatInt(int64(*v), 10) }
func (v *int64Value) IsZero(s string) bool { return s == "0" }
func (v *int64Value) Set(s string) (err error) {
	var res int64 // default `0` if input is empty
	if s != "" {
		res, err = strconv.ParseInt(s, 0, 64)
	}
	*v = int64Value(res)
	return err
}

// -- uint Value
type uintValue uint

func (v *uintValue) Type() string         { return "uint" }
func (v *uintValue) String() string       { return strconv.FormatUint(uint64(*v), 10) }
func (v *uintValue) IsZero(s string) bool { return s == "0" }
func (v *uintValue) Set(s string) (err error) {
	var res uint64 // default `0` if input is empty
	if s != "" {
		res, err = strconv.ParseUint(s, 0, strconv.IntSize)
	}
	*v = uintValue(res)
	return err
}

// -- uint64 Value
type uint64Value uint64

func (v *uint64Value) Type() string         { return "uint64" }
func (v *uint64Value) String() string       { return strconv.FormatUint(uint64(*v), 10) }
func (v *uint64Value) IsZero(s string) bool { return s == "0" }
func (v *uint64Value) Set(s string) (err error) {
	var res uint64 // default `0` if input is empty
	if s != "" {
		res, err = strconv.ParseUint(s, 0, 64)
	}
	*v = uint64Value(res)
	return err
}

// -- string Value
type stringValue string

func (v *stringValue) Type() string         { return "string" }
func (v *stringValue) String() string       { return string(*v) }
func (v *stringValue) IsZero(s string) bool { return s == "" }
func (v *stringValue) Set(s string) error {
	*v = stringValue(s)
	return nil
}

// -- float64 Value
type float64Value float64

func (v *float64Value) Type() string         { return "float64" }
func (v *float64Value) String() string       { return strconv.FormatFloat(float64(*v), 'g', -1, 64) }
func (v *float64Value) IsZero(s string) bool { return s == "0" }
func (v *float64Value) Set(s string) (err error) {
	var res float64 // default `0` if input is empty
	if s != "" {
		res, err = strconv.ParseFloat(s, 64)
	}
	*v = float64Value(res)
	return err
}

// -- time.Duration Value
type durationValue time.Duration

func (v *durationValue) Type() string         { return "duration" }
func (v *durationValue) String() string       { return (*time.Duration)(v).String() }
func (v *durationValue) IsZero(s string) bool { return s == "0s" }
func (v *durationValue) Set(s string) (err error) {
	var res time.Duration // default `0` if input is empty
	if s != "" {
		res, err = time.ParseDuration(s)
	}
	*v = durationValue(res)
	return err
}

// -- bytes Value
type bytesValue []byte

func (v *bytesValue) Type() string         { return "bytes" }
func (v *bytesValue) String() string       { return base64.StdEncoding.EncodeToString([]byte(*v)) }
func (v *bytesValue) IsZero(s string) bool { return s == "" }
func (v *bytesValue) Set(s string) (err error) {
	var res []byte // default `nil` if input is empty
	if s != "" {
		res, err = base64.StdEncoding.DecodeString(s)
	}
	*v = bytesValue(res)
	return err
}
