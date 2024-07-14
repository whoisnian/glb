package config_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/whoisnian/glb/config"
)

type TagField struct {
	T0 int
	T1 int `flag:"t"`
	T2 int `flag:"tt,22"`
	T3 int `flag:",33,"`
	T4 int `flag:",,This is T4"`

	NT struct {
		T5 string `flag:"|t-5"`
		T6 string `flag:"|t-6|abc"`
		T7 string `flag:"|t-7|a,b,c|This is T7"`
		T8 string `flag:"||def|"`
		T9 string `flag:"|||This is T9"`
	}
}

var tagFieldResults = [][]string{
	// {"Name", "DefValue", "Usage"},
	{"t0", "0", ""},
	{"t", "0", ""},
	{"tt", "22", ""},
	{"t3", "33", ""},
	{"t4", "0", "This is T4"},
	{"t-5", "", ""},
	{"t-6", "abc", ""},
	{"t-7", "a,b,c", "This is T7"},
	{"t8", "def", ""},
	{"t9", "", "This is T9"},
}

func TestNewFlagSet_ParseTagField(t *testing.T) {
	f, err := config.NewFlagSet(&TagField{})
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	for _, result := range tagFieldResults {
		flg := f.Lookup(result[0])
		if flg == nil {
			t.Fatalf("f.Lookup(%q) = nil, want *flag.Flag", result[0])
		}
		if flg.Name != result[0] || flg.DefValue != result[1] || flg.Usage != result[2] {
			t.Fatalf("f.Lookup(%q) = %q, want %q", result[0], []string{flg.Name, flg.DefValue, flg.Usage}, result)
		}
	}
}

type TagValue struct {
	Bool     bool          `flag:"bool,true,Enable feature xx"`
	Int      int           `flag:"int,0,Count of xx0"`
	Int64    int64         `flag:"int64,1,Count of xx1"`
	Uint     uint          `flag:"uint,2,Count of xx2"`
	Uint64   uint64        `flag:"uint64,3,Count of xx3"`
	String   string        `flag:"string,:80,Listen addr"`
	Float64  float64       `flag:"float64,0.6,Threshold of xx"`
	Duration time.Duration `flag:"duration,10s,Heartbeat interval"`
	Bytes    []byte        `flag:"bytes,d2hvaXNuaWFu,Private key (base64)"`
}

var tagValueResults = [][]string{
	// {"Name", "DefValue", "Usage"},
	{"bool", "true", "Enable feature xx"},
	{"int", "0", "Count of xx0"},
	{"int64", "1", "Count of xx1"},
	{"uint", "2", "Count of xx2"},
	{"uint64", "3", "Count of xx3"},
	{"string", ":80", "Listen addr"},
	{"float64", "0.6", "Threshold of xx"},
	{"duration", "10s", "Heartbeat interval"},
	{"bytes", "d2hvaXNuaWFu", "Private key (base64)"},
}

var tagValueUsage = `  -help     bool     Show usage message and quit
  -config   string   Specify file path of custom configuration json
  -bool     bool     Enable feature xx [CFG_BOOL] (default true)
  -int      int      Count of xx0 [CFG_INT]
  -int64    int64    Count of xx1 [CFG_INT64] (default 1)
  -uint     uint     Count of xx2 [CFG_UINT] (default 2)
  -uint64   uint64   Count of xx3 [CFG_UINT64] (default 3)
  -string   string   Listen addr [CFG_STRING] (default ":80")
  -float64  float64  Threshold of xx [CFG_FLOAT64] (default 0.6)
  -duration duration Heartbeat interval [CFG_DURATION] (default 10s)
  -bytes    bytes    Private key (base64) [CFG_BYTES] (default d2hvaXNuaWFu)
`

func TestNewFlagSet_ParseTagValue(t *testing.T) {
	f, err := config.NewFlagSet(&TagValue{})
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	for _, result := range tagValueResults {
		flg := f.Lookup(result[0])
		if flg == nil {
			t.Fatalf("f.Lookup(%q) = nil, want *flag.Flag", result[0])
		}
		if flg.Name != result[0] || flg.DefValue != result[1] || flg.Usage != result[2] {
			t.Fatalf("f.Lookup(%q) = %q, want %q", result[0], []string{flg.Name, flg.DefValue, flg.Usage}, result)
		}
	}
}

func TestNewFlagSet_FillDefaultValue(t *testing.T) {
	actual := TagValue{}
	want := TagValue{
		Bool:     true,
		Int:      0,
		Int64:    1,
		Uint:     2,
		Uint64:   3,
		String:   ":80",
		Float64:  0.6,
		Duration: time.Second * 10,
		Bytes:    []byte("whoisnian"),
	}
	_, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("config.NewFlagSet() result:\n  get  %+v\n  want %+v", actual, want)
	}
}

func TestNewFlagSet_TypeError(t *testing.T) {
	_, err := config.NewFlagSet([]int{1, 2, 3})
	if err == nil || err.Error() != "config: NewFlagSet() want pointer as input argument, but got slice" {
		t.Fatalf("config.NewFlagSet() = %q, want 'pointer as input argument' error", err)
	}
	_, err = config.NewFlagSet(&[]int{1, 2, 3})
	if err == nil || err.Error() != "config: NewFlagSet() want pointer to struct, but got pointer to slice" {
		t.Fatalf("config.NewFlagSet() = %q, want 'pointer to struct' error", err)
	}
	_, err = config.NewFlagSet(&struct{ f1, F2 float32 }{})
	if err == nil || err.Error() != "config: unknown value type float32" {
		t.Fatalf("config.NewFlagSet() = %q, want 'unknown value type float32' error", err)
	}
}

func TestNewFlagSet_FlagNameError(t *testing.T) {
	_, err := config.NewFlagSet(&struct {
		T int `flag:"-t"`
	}{})
	if err == nil || err.Error() != "config: flag name begins with -: -t" {
		t.Fatalf("config.NewFlagSet() = %q, want 'flag name begins with -' error", err)
	}
	_, err = config.NewFlagSet(&struct {
		T int `flag:"t="`
	}{})
	if err == nil || err.Error() != "config: flag name contains =: t=" {
		t.Fatalf("config.NewFlagSet() = %q, want 'flag name contains =' error", err)
	}
	_, err = config.NewFlagSet(&struct {
		T  int `flag:"tt"`
		NT struct {
			T int `flag:"tt"`
		}
	}{})
	if err == nil || err.Error() != "config: flag name redefined: tt" {
		t.Fatalf("config.NewFlagSet() = %q, want 'flag name redefined' error", err)
	}
}

func TestFromCommandLine(t *testing.T) {
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()
	os.Args = []string{"cmd", "--", "arg1", "arg2"}

	args, err := config.FromCommandLine(&struct{}{})
	if err != nil {
		t.Fatalf("config.FromCommandLine() error: %v", err)
	}
	want := []string{"arg1", "arg2"}
	if !reflect.DeepEqual(args, want) {
		t.Fatalf("config.FromCommandLine() result:\n  get  %+v\n  want %+v", args, want)
	}
}

func TestParse_Cli(t *testing.T) {
	arguments := []string{
		"-bool=false",
		"-int=10",
		"-uint64", "20",
		"--string=127.0.0.1:80",
		"--bytes", "X25pYW5f",
	}
	actual := TagValue{}
	want := TagValue{
		Bool:     false,
		Int:      10,
		Int64:    1, // default flag value
		Uint:     2, // default flag value
		Uint64:   20,
		String:   "127.0.0.1:80",
		Float64:  0.6,              // default flag value
		Duration: time.Second * 10, // default flag value
		Bytes:    []byte("_nian_"),
	}
	f, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if err := f.Parse(arguments); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("f.Parse() result:\n  get  %+v\n  want %+v", actual, want)
	}
}

func TestParse_Env(t *testing.T) {
	arguments := []string{}
	actual := TagValue{}
	want := TagValue{
		Bool:     false,
		Int:      10,
		Int64:    -20,
		Uint:     2,     // default flag value
		Uint64:   3,     // default flag value
		String:   ":80", // default flag value
		Float64:  0.01,
		Duration: time.Minute * 5,
		Bytes:    []byte("whoisnian"), // default flag value
	}
	f, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}

	os.Setenv("CFG_BOOL", "false")
	os.Setenv("CFG_INT", "10")
	os.Setenv("CFG_INT64", "-20")
	os.Setenv("CFG_FLOAT64", "0.01")
	os.Setenv("CFG_DURATION", "5m")

	if err := f.Parse(arguments); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("f.Parse() result:\n  get  %+v\n  want %+v", actual, want)
	}

	os.Unsetenv("CFG_BOOL")
	os.Unsetenv("CFG_INT")
	os.Unsetenv("CFG_INT64")
	os.Unsetenv("CFG_FLOAT64")
	os.Unsetenv("CFG_DURATION")
}

func TestParse_Error(t *testing.T) {
	f, err := config.NewFlagSet(&TagValue{})
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if err := f.Parse([]string{}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if err := f.Parse([]string{}); err == nil || err.Error() != "config: Parse() must be called once" {
		t.Fatalf("f.Parse() = %q, want 'must be called once' error", err)
	}
}

func TestArgs(t *testing.T) {
	var tests = [][][]string{{
		{"arg_1"},
		{"arg_1"},
	}, {
		{"arg_1", "arg_2"},
		{"arg_1", "arg_2"},
	}, {
		{"-int=10", "arg_1", "arg_2"},
		{"arg_1", "arg_2"},
	}, {
		{"arg_1", "-int=10", "arg_3"},
		{"arg_1", "-int=10", "arg_3"},
	}}
	for _, tt := range tests {
		f, err := config.NewFlagSet(&TagValue{})
		if err != nil {
			t.Fatalf("config.NewFlagSet() error: %v", err)
		}
		if err := f.Parse(tt[0]); err != nil {
			t.Fatalf("f.Parse() error: %v", err)
		}
		if !reflect.DeepEqual(f.Args(), tt[1]) {
			t.Fatalf("f.Args() result:\n  get  %+v\n  want %+v", f.Args(), tt[1])
		}
	}
}

func TestConfigJson_File(t *testing.T) {
	actual := TagValue{}
	want := TagValue{
		Bool:     false,
		Int:      10,
		Int64:    -20,
		Uint:     2,
		Uint64:   3,
		String:   "0.0.0.0:80",
		Float64:  0.01,
		Duration: time.Minute * 5,
		Bytes:    []byte("_nian_"),
	}

	fi, err := os.CreateTemp("", "config-json-file-*.json")
	if err != nil {
		t.Fatalf("os.CreateTemp() error: %v", err)
	}
	if err = json.NewEncoder(fi).Encode(want); err != nil {
		t.Fatalf("json.Encode() error: %v", err)
	}
	fi.Close()
	defer os.Remove(fi.Name())

	f, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if err := f.Parse([]string{"-config", fi.Name()}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("f.Parse() result:\n  get  %+v\n  want %+v", actual, want)
	}
}

func TestConfigJson_Env(t *testing.T) {
	actual := TagValue{}
	want := TagValue{
		Bool:     false,
		Int:      10,
		Int64:    -20,
		Uint:     2,
		Uint64:   3,
		String:   "0.0.0.0:80",
		Float64:  0.01,
		Duration: time.Minute * 5,
		Bytes:    []byte("_nian_"),
	}

	data, err := json.Marshal(want)
	if err != nil {
		t.Fatalf("json.Marshal() error: %v", err)
	}
	os.Setenv("CFG_CONFIG_B64", base64.RawStdEncoding.EncodeToString(data))

	f, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if err := f.Parse(nil); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("f.Parse() result:\n  get  %+v\n  want %+v", actual, want)
	}
	os.Unsetenv("CFG_CONFIG_B64")
}

func TestValuePriority(t *testing.T) {
	actual := TagValue{}
	want := TagValue{
		Bool:     true,
		Int:      12, // cli > env > file > default
		Int64:    21, // env > file > default
		Uint:     30, // file > default
		Uint64:   3,  // default
		String:   ":80",
		Float64:  0.6,
		Duration: time.Second * 10,
		Bytes:    []byte("whoisnian"),
	}

	fi, err := os.CreateTemp("", "config-json-file-*.json")
	if err != nil {
		t.Fatalf("os.CreateTemp() error: %v", err)
	}
	if _, err = fi.WriteString(`{"Int":10,"Int64":20,"Uint":30}`); err != nil {
		t.Fatalf("File.WriteString() error: %v", err)
	}
	fi.Close()
	defer os.Remove(fi.Name())

	f, err := config.NewFlagSet(&actual)
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}

	os.Setenv("CFG_INT", "11")
	os.Setenv("CFG_INT64", "21")

	if err := f.Parse([]string{"-config", fi.Name(), "-int=12"}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("f.Parse() result:\n  get  %+v\n  want %+v", actual, want)
	}

	os.Unsetenv("CFG_INT")
	os.Unsetenv("CFG_INT64")
}

func TestPrintUsage(t *testing.T) {
	f, err := config.NewFlagSet(&TagValue{})
	if err != nil {
		t.Fatalf("config.NewFlagSet() error: %v", err)
	}
	if err := f.Parse([]string{"-help"}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if !f.ShowUsage() {
		t.Fatalf("f.ShowUsage() = %v, want true", f.ShowUsage())
	}
	out := &bytes.Buffer{}
	f.PrintUsage(out, false)
	if out.String() != tagValueUsage {
		t.Fatalf("f.PrintUsage() result:\nget:\n%swant:\n%s", out.String(), tagValueUsage)
	}
}
