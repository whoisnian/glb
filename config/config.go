// Package config generates flag.FlagSet from structField tags and reads Flag.Value from cli, env, or configuration file.
//
// priority: cli > env > file > default
//
// Two structField tag formats:
//
//	`flag:"name,value,usage"`
//	`flag:"|name|value|usage"`
//
// Nested structure is not supported. Type of structField can be:
//   - *bool
//   - *int
//   - *int64
//   - *uint
//   - *uint64
//   - *string
//   - *float64
//   - *time.Duration
//   - *[]byte
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

type FlagSet struct {
	set       *flag.FlagSet
	nameMap   map[string]string     // mapping from flag name to original struct field name
	actualMap map[string]*flag.Flag // flags in the internal flagSet that have been set by cli, env, or custom configuration

	configSource string
	configB64Env string
	envPrefix    string

	initialized bool
	structValue reflect.Value
}

// NewFlagSet returns a new, empty flag set with the specified name and error handling property.
func NewFlagSet(name string, errorHandling flag.ErrorHandling) *FlagSet {
	f := &FlagSet{
		set:          flag.NewFlagSet(name, errorHandling),
		nameMap:      make(map[string]string),
		actualMap:    make(map[string]*flag.Flag),
		configB64Env: "CFG_CONFIG_B64",
		envPrefix:    "CFG_",
	}

	// cli usage:
	//   -config=config.json
	//   -config ./config.json
	//   --config=~/config.json
	//   --config /etc/demo/config.json
	f.set.StringVar(&f.configSource, "config", "", "Specify file path of custom configuration json")
	return f
}

// LookupActual returns the Flag structure that have been set, returning nil if not set.
func (f *FlagSet) LookupActual(name string) *flag.Flag {
	return f.actualMap[name]
}

// LookupFormal returns the Flag structure of the named flag, returning nil if none exists.
func (f *FlagSet) LookupFormal(name string) *flag.Flag {
	return f.set.Lookup(name)
}

// Args returns the non-flag command-line arguments, similarly to `flag.Args()`.
func (f *FlagSet) Args() []string {
	return f.set.Args()
}

// Initialized reports whether f.Init() has been called.
func (f *FlagSet) Initialized() bool {
	return f.initialized
}

// FromCommandLine creates new flag set and parses os.Args for input struct argument.
//
// Example:
//
//	type Config struct {
//	    ListenAddr string `flag:"l,0.0.0.0:80,Server listen addr"`
//	    Version    bool   `flag:"v,false,Show version and quit"`
//	}
//
//	func main() {
//	    cfg := &Config{}
//	    if err := config.FromCommandLine(cfg); err != nil {
//	        panic(err)
//	    }
//	    fmt.Printf("%+v", cfg)
//	}
func FromCommandLine(pStruct any) (err error) {
	f := NewFlagSet(os.Args[0], flag.ExitOnError)
	if err = f.Init(pStruct); err != nil {
		return err
	}
	if err = f.Parse(os.Args[1:]); err != nil {
		return err
	}
	return nil
}

// FromCommandLineWithArgs parses os.Args for input struct argument and returns the non-flag arguments.
func FromCommandLineWithArgs(pStruct any) (args []string, err error) {
	f := NewFlagSet(os.Args[0], flag.ExitOnError)
	if err = f.Init(pStruct); err != nil {
		return nil, err
	}
	if err = f.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	return f.Args(), nil
}

// GenerateDefault fills the value of the pStruct pointing to using default values.
func GenerateDefault(pStruct any) (err error) {
	f := NewFlagSet("", flag.ContinueOnError)
	if err = f.Init(pStruct); err != nil {
		return err
	}
	f.set.VisitAll(func(flg *flag.Flag) {
		if err == nil {
			err = flg.Value.Set(flg.DefValue)
		}
	})
	return err
}

// Init parses struct exported fields as configuration items and adds items to the internal flagSet.
// Only accept a pointer to struct as input argument.
func (f *FlagSet) Init(pStruct any) error {
	if f.initialized {
		return errors.New("config: Init() should be called only once")
	}
	f.initialized = true

	pVal := reflect.ValueOf(pStruct)
	if pVal.Kind() != reflect.Pointer {
		return errors.New("config: Init() want pointer as input argument, but got " + pVal.Kind().String())
	}
	f.structValue = pVal.Elem()
	if f.structValue.Kind() != reflect.Struct {
		return errors.New("config: Init() want pointer to struct, but got pointer to " + f.structValue.Kind().String())
	}

	vType := f.structValue.Type()
	for i := 0; i < vType.NumField(); i++ {
		sf := vType.Field(i)
		if !sf.IsExported() {
			continue
		}

		tName, tValue, tUsage := parseStructFieldTag(sf)
		f.nameMap[tName] = sf.Name
		if err := f.addFieldToFlagSet(i, tName, tValue, tUsage); err != nil {
			return err
		}
	}
	return nil
}

// Parse parses flag definitions from the argument list, which should not include the command name.
// Must be called after f.Init().
func (f *FlagSet) Parse(arguments []string) error {
	if !f.initialized {
		return errors.New("config: Parse() must be called after f.Init()")
	}
	f.set.Parse(arguments)

	configJsonMap := make(map[string]json.RawMessage)
	err := f.unmarshalConfigJson(&configJsonMap)
	if err != nil {
		return err
	}

	f.set.Visit(func(flg *flag.Flag) { f.actualMap[flg.Name] = flg })
	f.set.VisitAll(func(flg *flag.Flag) {
		if f.actualMap[flg.Name] != nil {
			return // flag has been set by cli
		}

		if res, ok := os.LookupEnv(f.envPrefix + strings.ToUpper(f.nameMap[flg.Name])); ok {
			flg.Value.Set(res)
			f.actualMap[flg.Name] = flg
			return // flag has been set by env
		}

		if res, ok := configJsonMap[f.nameMap[flg.Name]]; ok {
			json.Unmarshal(res, f.structValue.FieldByName(f.nameMap[flg.Name]).Addr().Interface())
			f.actualMap[flg.Name] = flg
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
	name = sf.Tag.Get("flag")

	var sep byte = ','
	if name != "" && name[0] == '|' {
		name = name[1:]
		sep = '|'
	}

	if posN := strings.IndexByte(name, sep); posN >= 0 {
		value = name[posN+1:]
		name = name[:posN]
		if posV := strings.IndexByte(value, sep); posV >= 0 {
			usage = value[posV+1:]
			value = value[:posV]
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
func (f *FlagSet) addFieldToFlagSet(i int, name string, defValue string, usage string) error {
	switch pVar := f.structValue.Field(i).Addr().Interface().(type) {
	case *bool:
		value, err := parseDefaultBool(defValue)
		if err != nil {
			return err
		}
		f.set.BoolVar(pVar, name, value, usage)
	case *int:
		value, err := parseDefaultInt(defValue)
		if err != nil {
			return err
		}
		f.set.IntVar(pVar, name, value, usage)
	case *int64:
		value, err := parseDefaultInt64(defValue)
		if err != nil {
			return err
		}
		f.set.Int64Var(pVar, name, value, usage)
	case *uint:
		value, err := parseDefaultUint(defValue)
		if err != nil {
			return err
		}
		f.set.UintVar(pVar, name, value, usage)
	case *uint64:
		value, err := parseDefaultUint64(defValue)
		if err != nil {
			return err
		}
		f.set.Uint64Var(pVar, name, value, usage)
	case *string:
		f.set.StringVar(pVar, name, defValue, usage)
	case *float64:
		value, err := parseDefaultFloat64(defValue)
		if err != nil {
			return err
		}
		f.set.Float64Var(pVar, name, value, usage)
	case *time.Duration:
		value, err := parseDefaultDuration(defValue)
		if err != nil {
			return err
		}
		f.set.DurationVar(pVar, name, value, usage)
	case *[]byte:
		value, err := parseDefaultBytesValue(defValue)
		if err != nil {
			return err
		}
		f.set.TextVar((*bytesValue)(pVar), name, value, usage)
	default:
		return errors.New("config: unknown var type " + reflect.TypeOf(pVar).String())
	}
	return nil
}

func (f *FlagSet) unmarshalConfigJson(configJsonMap *map[string]json.RawMessage) (err error) {
	var data []byte
	if f.configSource != "" {
		fPath, err := fsutil.ExpandHomeDir(f.configSource)
		if err != nil {
			return err
		}
		data, err = os.ReadFile(fPath)
		if err != nil {
			return err
		}
	} else if res, ok := os.LookupEnv(f.configB64Env); ok {
		data, err = base64.StdEncoding.DecodeString(res)
		if err != nil {
			return err
		}
	} else {
		return nil
	}
	return json.Unmarshal(data, &configJsonMap)
}
