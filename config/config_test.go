package config_test

import (
	"encoding/base64"
	"encoding/json"
	"flag"
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

	T5 string `flag:"|t-5"`
	T6 string `flag:"|t-6|abc"`
	T7 string `flag:"|t-7|a,b,c|This is T7"`
	T8 string `flag:"||def|"`
	T9 string `flag:"|||This is T9"`
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

func TestInit_ParseTagField(t *testing.T) {
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&TagField{}); err != nil {
		t.Fatalf("f.Init() error: %v", err)
	}
	for _, result := range tagFieldResults {
		flg := f.LookupFormal(result[0])
		if flg == nil {
			t.Fatalf("f.LookupFormal(%q) = nil, want *flag.Flag", result[0])
		}
		if flg.Name != result[0] || flg.DefValue != result[1] || flg.Usage != result[2] {
			t.Fatalf("f.LookupFormal(%q) = %q, want %q", result[0], []string{flg.Name, flg.DefValue, flg.Usage}, result)
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

func TestInit_ParseTagValue(t *testing.T) {
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&TagValue{}); err != nil {
		t.Fatalf("f.Init() error: %v", err)
	}
	for _, result := range tagValueResults {
		flg := f.LookupFormal(result[0])
		if flg == nil {
			t.Fatalf("f.LookupFormal(%q) = nil, want *flag.Flag", result[0])
		}
		if flg.Name != result[0] || flg.DefValue != result[1] || flg.Usage != result[2] {
			t.Fatalf("f.LookupFormal(%q) = %q, want %q", result[0], []string{flg.Name, flg.DefValue, flg.Usage}, result)
		}
	}
}

func TestInit_DuplicatedCall(t *testing.T) {
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if f.Initialized() {
		t.Fatal("f.Initialized() = true, want false")
	}
	if err := f.Init(&TagValue{}); err != nil {
		t.Fatalf("f.Init() error: %v", err)
	}
	if !f.Initialized() {
		t.Fatal("f.Initialized() = false, want true")
	}
	if err := f.Init(&TagValue{}); err == nil || err.Error() != "config: Init() should be called only once" {
		t.Fatalf("f.Init() = %q, want 'called only once' error", err)
	}
}

func TestInit_TypeError(t *testing.T) {
	err := config.NewFlagSet("test", flag.ContinueOnError).Init([]int{1, 2, 3})
	if err == nil || err.Error() != "config: Init() want pointer as input argument, but got slice" {
		t.Fatalf("f.Init() = %q, want 'pointer as input argument' error", err)
	}
	err = config.NewFlagSet("test", flag.ContinueOnError).Init(&[]int{1, 2, 3})
	if err == nil || err.Error() != "config: Init() want pointer to struct, but got pointer to slice" {
		t.Fatalf("f.Init() = %q, want 'pointer to struct' error", err)
	}
	err = config.NewFlagSet("test", flag.ContinueOnError).Init(&struct{ f1, F2 float32 }{})
	if err == nil || err.Error() != "config: unknown var type *float32" {
		t.Fatalf("f.Init() = %q, want 'unknown var type' error", err)
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
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&actual); err != nil {
		t.Fatalf("f.Init() error: %v", err)
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
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&actual); err != nil {
		t.Fatalf("f.Init() error: %v", err)
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
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Parse([]string{}); err == nil || err.Error() != "config: Parse() must be called after f.Init()" {
		t.Fatalf("f.Parse() = %q, want 'must be called after' error", err)
	}
	if err := f.Init(&TagValue{}); err != nil {
		t.Fatalf("f.Init() error: %v", err)
	}
	if err := f.Parse([]string{}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
}

func TestLookupActual(t *testing.T) {
	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&TagValue{}); err != nil {
		t.Fatalf("f.Init() error: %v", err)
	}
	if err := f.Parse([]string{"-int=10", "-uint=20"}); err != nil {
		t.Fatalf("f.Parse() error: %v", err)
	}
	if res := f.LookupActual("bool"); res != nil {
		t.Fatalf("f.LookupActual('bool') = %q, want nil", res)
	}
	if res := f.LookupActual("bytes"); res != nil {
		t.Fatalf("f.LookupActual('bytes') = %q, want nil", res)
	}
	if res := f.LookupActual("int"); res == nil || res.Value.String() != "10" {
		t.Fatalf("f.LookupActual('int') = %q, want 10", res.Value)
	}
	if res := f.LookupActual("uint"); res == nil || res.Value.String() != "20" {
		t.Fatalf("f.LookupActual('uint') = %q, want 20", res.Value)
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
		f := config.NewFlagSet("test", flag.ContinueOnError)
		if err := f.Init(&TagValue{}); err != nil {
			t.Fatalf("f.Init() error: %v", err)
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

	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&actual); err != nil {
		t.Fatalf("f.Init() error: %v", err)
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

	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&actual); err != nil {
		t.Fatalf("f.Init() error: %v", err)
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

	f := config.NewFlagSet("test", flag.ContinueOnError)
	if err := f.Init(&actual); err != nil {
		t.Fatalf("f.Init() error: %v", err)
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

func TestGenerateExample(t *testing.T) {
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
	err := config.GenerateDefault(&actual)
	if err != nil {
		t.Fatalf("config.GenerateExample() error: %v", err)
	}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("config.GenerateExample() result:\n  get  %+v\n  want %+v", actual, want)
	}
}
