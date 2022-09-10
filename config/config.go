// Package config generates flag.FlagSet from structField tags and reads Flag.Value from cli, env, or configuration file.
//
// priority: cli > env > file > default
package config

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/whoisnian/glb/util/fsutil"
)

var (
	flagSet   = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	nameMap   = make(map[string]string)     // mapping from flag name to original struct field name
	actualMap = make(map[string]*flag.Flag) // flags in flagSet that have been set by cli, env, or custom configuration

	configSource string
	configB64Env = "CFG_CONFIG_B64"
	envPrefix    = "CFG_"
)

// LookupActual returns the Flag structure that have been set, returning nil if not set.
func LookupActual(name string) *flag.Flag {
	return actualMap[name]
}

// LookupFormal returns the Flag structure of the named flag, returning nil if none exists.
func LookupFormal(name string) *flag.Flag {
	return flagSet.Lookup(name)
}

// Init parses struct exported fields as configuration items and fills final values into the struct.
// This function accepts a pointer to struct as input argument.
func Init(pStruct any) error {
	pVal := reflect.ValueOf(pStruct)
	if pVal.Kind() != reflect.Pointer {
		return errors.New("config: Init() want pointer as input argument, but got " + pVal.Kind().String())
	}
	val := pVal.Elem()
	if val.Kind() != reflect.Struct {
		return errors.New("config: Init() want pointer to struct, but got pointer to " + val.Kind().String())
	}

	// cli usage:
	//   -config=config.json
	//   -config ./config.json
	//   --config=~/config.json
	//   --config /etc/demo/config.json
	flagSet.StringVar(&configSource, "config", "", "Specify file path of custom configuration json")

	vType := val.Type()
	for i := 0; i < vType.NumField(); i++ {
		sf := vType.Field(i)
		if !sf.IsExported() {
			continue
		}

		tName, tValue, tUsage := parseStructFieldTag(sf)
		nameMap[tName] = sf.Name
		if err := addVarToFlagSet(val.Field(i).Addr().Interface(), tName, tValue, tUsage); err != nil {
			return err
		}
	}
	flagSet.Parse(os.Args[1:])

	configJsonMap := make(map[string]json.RawMessage)
	err := unmarshalConfigJson(&configJsonMap)
	if err != nil {
		return err
	}

	flagSet.Visit(func(f *flag.Flag) { actualMap[f.Name] = f })
	flagSet.VisitAll(func(f *flag.Flag) {
		if actualMap[f.Name] != nil {
			return // flag has been set by cli
		}

		if res, ok := os.LookupEnv(envPrefix + strings.ToUpper(nameMap[f.Name])); ok {
			f.Value.Set(res)
			actualMap[f.Name] = f
			return // flag has been set by env
		}

		if res, ok := configJsonMap[nameMap[f.Name]]; ok {
			json.Unmarshal(res, val.FieldByName(nameMap[f.Name]).Addr().Interface())
			actualMap[f.Name] = f
			return // flag has been set by custom configuration
		}
	})

	return nil
}

// tag format:
//
//	`flag:"name,value,usage"`
//	`flag:"|name|value|usage"`
func parseStructFieldTag(sf reflect.StructField) (name, value, usage string) {
	src := sf.Tag.Get("flag")

	var sep byte = ','
	if src != "" && src[0] == '|' {
		src = src[1:]
		sep = '|'
	}

	if posN := strings.IndexByte(src, sep); posN >= 0 {
		name = src[:posN]
		if posV := strings.IndexByte(src[posN+1:], sep); posV >= 0 {
			value = src[posN+1 : posN+1+posV]
			usage = src[posN+1+posV+1:]
		}
	}
	if name == "" {
		name = strings.ToLower(sf.Name)
	}
	return name, value, usage
}

// Type of variable can be:
//   - *bool
//   - *int
//   - *int64
//   - *uint
//   - *uint64
//   - *string
//   - *float64
//   - *time.Duration
//   - *[]byte
func addVarToFlagSet(variable any, name string, defValue string, usage string) error {
	switch pVar := variable.(type) {
	case *bool:
		value, err := parseDefaultBool(defValue)
		if err != nil {
			return err
		}
		flagSet.BoolVar(pVar, name, value, usage)
	case *int:
		value, err := parseDefaultInt(defValue)
		if err != nil {
			return err
		}
		flagSet.IntVar(pVar, name, value, usage)
	case *int64:
		value, err := parseDefaultInt64(defValue)
		if err != nil {
			return err
		}
		flagSet.Int64Var(pVar, name, value, usage)
	case *uint:
		value, err := parseDefaultUint(defValue)
		if err != nil {
			return err
		}
		flagSet.UintVar(pVar, name, value, usage)
	case *uint64:
		value, err := parseDefaultUint64(defValue)
		if err != nil {
			return err
		}
		flagSet.Uint64Var(pVar, name, value, usage)
	case *string:
		flagSet.StringVar(pVar, name, defValue, usage)
	case *float64:
		value, err := parseDefaultFloat64(defValue)
		if err != nil {
			return err
		}
		flagSet.Float64Var(pVar, name, value, usage)
	case *time.Duration:
		value, err := parseDefaultDuration(defValue)
		if err != nil {
			return err
		}
		flagSet.DurationVar(pVar, name, value, usage)
	case *[]byte:
		value, err := parseDefaultBytesValue(defValue)
		if err != nil {
			return err
		}
		flagSet.TextVar((*bytesValue)(pVar), name, value, usage)
	default:
		return errors.New("config: unknown var type " + reflect.TypeOf(pVar).String())
	}
	return nil
}

func unmarshalConfigJson(configJsonMap *map[string]json.RawMessage) (err error) {
	var data []byte
	if configSource != "" {
		fPath, err := fsutil.ResolveHomeDir(configSource)
		if err != nil {
			return err
		}
		data, err = os.ReadFile(fPath)
		if err != nil {
			return err
		}
	} else if res, ok := os.LookupEnv(configB64Env); ok {
		data, err = base64.StdEncoding.DecodeString(res)
		if err != nil {
			return err
		}
	} else {
		return nil
	}
	return json.Unmarshal(data, &configJsonMap)
}
