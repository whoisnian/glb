// Package config generates flags from StructField tags and reads values from cli, env, and configuration file.
//
// priority: cli > env > file > default
//
// Two StructField tag formats are supported. Extra separators will be treated as usage content.
//
//	`flag:"name,value,usage"`
//	`flag:"|name|value|usage"`
//
// If the tag is empty, the lowercase StructField name will be used. Type of StructField could be:
//   - bool
//   - int
//   - int64
//   - uint
//   - uint64
//   - string
//   - float64
//   - time.Duration
//   - []byte
//   - struct
package config

import (
	"encoding/base64"
	"errors"
	"os"
	"reflect"
	"strings"

	"github.com/whoisnian/glb/util/fsutil"
)

const (
	flagNameShowUsage  = "help"   // Show usage message and quit
	flagNameConfigPath = "config" // Specify file path of custom configuration json
)

// FlagSet represents a set of defined flags. Flag names must be unique within a FlagSet.
type FlagSet struct {
	ptr  any      // pointer to config struct
	args []string // arguments after flags

	flagList  []*Flag          // record flags order
	flagMap   map[string]*Flag // lookup a flag by name
	maxLength int              // flag name max length

	parsed       bool
	envKeyPrefix string
	b64ConfigEnv string

	// built-in flag values
	valueShowUsage  boolValue
	valueConfigPath stringValue
}

// Flag represents the state of a flag.
type Flag struct {
	Name     string // name as it appears on command line
	Env      string // environment variable name
	Usage    string // help message
	Value    Value  // value as set
	DefValue string // default value (as text); for usage message

	ArgValue *string // command line argument value (as text, nil if not set); for argParse()
	EnvValue *string // environment variable value (as text, nil if not set); for envParse()
}

// NewFlagSet parses struct exported fields and generates a new FlagSet.
// Only accepts a pointer to struct as input argument.
func NewFlagSet(pStruct any) (*FlagSet, error) {
	pVal := reflect.ValueOf(pStruct)
	if pVal.Kind() != reflect.Pointer {
		return nil, errors.New("config: NewFlagSet() want pointer as input argument, but got " + pVal.Kind().String())
	}
	structValue := pVal.Elem()
	if structValue.Kind() != reflect.Struct {
		return nil, errors.New("config: NewFlagSet() want pointer to struct, but got pointer to " + structValue.Kind().String())
	}

	f := &FlagSet{
		ptr:  pStruct,
		args: []string{},

		flagList:  []*Flag{},
		flagMap:   make(map[string]*Flag),
		maxLength: 0,

		parsed:       false,
		envKeyPrefix: "CFG_",
		b64ConfigEnv: "CFG_CONFIG_B64",
	}
	for _, flg := range []*Flag{
		{Name: flagNameShowUsage, Usage: "Show usage message and quit", Value: &f.valueShowUsage},
		{Name: flagNameConfigPath, Usage: "Specify file path of custom configuration json", Value: &f.valueConfigPath},
	} {
		f.flagList = append(f.flagList, flg)
		f.flagMap[flg.Name] = flg
		f.maxLength = max(f.maxLength, len(flg.Name))
	}
	return f, f.parseStructFields(structValue, "")
}

// Type of StructField could be:
//   - bool
//   - int
//   - int64
//   - uint
//   - uint64
//   - string
//   - float64
//   - time.Duration
//   - []byte
//   - struct
func (f *FlagSet) parseStructFields(structValue reflect.Value, group string) error {
	structType := structValue.Type()
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if !field.IsExported() {
			continue
		}

		fieldValue := structValue.Field(i)
		if field.Type.Kind() == reflect.Struct {
			if err := f.parseStructFields(fieldValue, group+field.Name+"_"); err != nil {
				return err
			}
			continue
		}

		name, defValue, usage := parseStructFieldTag(field)
		if strings.HasPrefix(name, "-") {
			return errors.New("config: flag name begins with -: " + name)
		} else if strings.Contains(name, "=") {
			return errors.New("config: flag name contains =: " + name)
		} else if _, ok := f.flagMap[name]; ok {
			return errors.New("config: flag name redefined: " + name)
		}

		value, err := newFlagValue(fieldValue, defValue)
		if err != nil {
			return err
		}

		flg := &Flag{
			Name:     name,
			Env:      strings.ToUpper(f.envKeyPrefix + group + field.Name),
			Usage:    usage,
			Value:    value,
			DefValue: value.String(), // after value.Set(defValue)
		}
		f.flagList = append(f.flagList, flg)
		f.flagMap[flg.Name] = flg
		f.maxLength = max(f.maxLength, len(flg.Name))
	}
	return nil
}

// Two StructField tag formats are supported. Extra separators will be treated as usage content.
//
//	`flag:"name,value,usage"`
//	`flag:"|name|value|usage"`
func parseStructFieldTag(field reflect.StructField) (name, value, usage string) {
	name = field.Tag.Get("flag")
	sep := byte(',')
	if name != "" && name[0] == '|' {
		name = name[1:]
		sep = '|'
	}

	if pos := strings.IndexByte(name, sep); pos >= 0 {
		value = name[pos+1:]
		name = name[:pos]
		if pos = strings.IndexByte(value, sep); pos >= 0 {
			usage = value[pos+1:]
			value = value[:pos]
		}
	}

	if name == "" {
		name = strings.ToLower(field.Name)
	}
	return name, value, usage
}

// Lookup returns the Flag structure of the named flag, returning nil if none exists.
func (f *FlagSet) Lookup(name string) *Flag {
	return f.flagMap[name]
}

// Args returns the non-flag command-line arguments, similarly to `flag.Args()`.
func (f *FlagSet) Args() []string {
	return f.args
}

// FromCommandLine creates new flag set and parses os.Args for input struct argument.
//
// Example:
//
//	type Config struct {
//	    Debug      bool   `flag:"d,false,Enable debug output"`
//	    Version    bool   `flag:"v,false,Show version and quit"`
//	    ListenAddr string `flag:"l,0.0.0.0:80,Server listen addr"`
//	}
//
//	func main() {
//	    var cfg Config
//	    if _, err := config.FromCommandLine(&cfg); err != nil {
//	        panic(err)
//	    }
//	    fmt.Printf("%+v", cfg)
//	}
func FromCommandLine(pStruct any) ([]string, error) {
	f, err := NewFlagSet(pStruct)
	if err != nil {
		return nil, err
	}
	if err = f.Parse(os.Args[1:]); err != nil {
		return nil, err
	}
	return f.Args(), nil
}

// Parse parses flag definitions from command line, environment variable, and configuration file.
// The command line argument list should not include the command name.
func (f *FlagSet) Parse(arguments []string) (err error) {
	if f.parsed {
		return errors.New("config: Parse() must be called once")
	}
	f.parsed = true
	f.args = arguments

	if err = f.argParse(); err != nil {
		return err
	}
	if err = f.envParse(); err != nil {
		return err
	}

	if flg, ok := f.flagMap[flagNameConfigPath]; ok && flg.ArgValue != nil {
		if err = flg.Value.Set(*flg.ArgValue); err != nil {
			return err
		}
	}
	if err = f.parseConfigJson(); err != nil {
		return err
	}

	for _, flg := range f.flagList {
		if flg.ArgValue != nil {
			err = flg.Value.Set(*flg.ArgValue)
		} else if flg.EnvValue != nil {
			err = flg.Value.Set(*flg.EnvValue)
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// The following forms are permitted.
//
//	-config=config.json
//	-config ./config.json
//	--config=~/config.json
//	--config /etc/demo/config.json
func (f *FlagSet) argParse() error {
	for len(f.args) > 0 {
		name := f.args[0]
		if len(name) < 2 || name[0] != '-' {
			return nil
		}
		name = name[1:]

		if name[0] == '-' {
			if len(name) == 2 {
				f.args = f.args[1:]
				return nil
			}
			name = name[1:]
		}
		if len(name) == 0 || name[0] == '-' || name[0] == '=' {
			return errors.New("config: bad flag syntax: " + name)
		}

		f.args = f.args[1:]
		hasValue := false
		argValue := ""
		for i := 1; i < len(name); i++ {
			if name[i] == '=' {
				argValue = name[i+1:]
				name = name[0:i]
				hasValue = true
				break
			}
		}

		flg, ok := f.flagMap[name]
		if !ok {
			return errors.New("config: flag provided but not defined: " + name)
		}

		if !hasValue {
			if fv, ok := flg.Value.(boolFlag); ok && fv.IsBoolFlag() {
				argValue = "true"
			} else if len(f.args) > 0 {
				argValue, f.args = f.args[0], f.args[1:]
			} else {
				return errors.New("config: flag needs an argument: " + name)
			}
		}
		flg.ArgValue = &argValue
	}
	return nil
}

func (f *FlagSet) envParse() error {
	for _, flg := range f.flagList {
		if flg.Env == "" {
			continue
		}
		// fmt.Println("LOOKUP " + flg.Name + " " + flg.Env)
		if res, ok := os.LookupEnv(flg.Env); ok {
			flg.EnvValue = &res
		}
	}
	return nil
}

func (f *FlagSet) parseConfigJson() (err error) {
	var jsonData []byte
	if fPath := f.valueConfigPath.String(); fPath != "" {
		if fPath, err = fsutil.ExpandHomeDir(fPath); err != nil {
			return err
		}
		if jsonData, err = os.ReadFile(fPath); err != nil {
			return err
		}
	} else if str, ok := os.LookupEnv(f.b64ConfigEnv); ok {
		if jsonData, err = base64.StdEncoding.DecodeString(str); err != nil {
			return err
		}
	} else {
		return nil
	}
	return JsonUnmarshal(jsonData, f.ptr)
}
