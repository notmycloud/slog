package nmcslog

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/invopop/jsonschema"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// FormatText specifies the usage of the slog Text output handler.
	FormatText OutputFormat = "TEXT"
	// FormatJSON specifies the usage of the slog JSON output handler.
	FormatJSON OutputFormat = "JSON"

	// LevelTrace defines the Trace Log Level (-8)
	LevelTrace = slog.LevelDebug - 4
	// LevelDebug defines the Debug Log Level (-4)
	LevelDebug = slog.LevelDebug
	// LevelInfo defines the Info Log Level (0)
	LevelInfo = slog.LevelInfo
	// LevelNotice defines the Notice Log Level (2)
	LevelNotice = slog.LevelInfo + 2
	// LevelWarn defines the Warn Log Level (4)
	LevelWarn = slog.LevelWarn
	// LevelError defines the Error Log Level (8)
	LevelError = slog.LevelError
	// LevelEmergency defines the Emergency Log Level (12)
	LevelEmergency = slog.LevelError + 4

	// DefaultRotateSize is the default max log filesize in Megabytes.
	DefaultRotateSize = 5
	// DefaultRotateKeep is the default number of log files to keep.
	DefaultRotateKeep = 4
	// DefaultRotateAge is the default max age of a log file in days.
	DefaultRotateAge = 7
)

var (
	ErrInvalidLogLevel = errors.New("invalid log level")
)

// LogLevel handles the decoding and parsing of a given log level to the slog log level.
type LogLevel struct {
	// Level to cutoff log messages, anything below this level will be dropped.
	Level string
	// Level is the converted Level from either an integer or string.
	level slog.Level
}

func (ll *LogLevel) DecodeLevel() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: decode level string %q: %w", ll.Level, err)
		}
	}()

	if ll.Level == "" {
		ll.level = slog.LevelInfo
		return nil
	}
	if l, err := strconv.ParseInt(ll.Level, 10, 64); err == nil {
		ll.level = slog.Level(l)
		return nil
	}

	name := ll.Level
	offset := 0
	if i := strings.IndexAny(ll.Level, "+-"); i >= 0 {
		name = ll.Level[:i]
		offset, err = strconv.Atoi(ll.Level[i:])
		if err != nil {
			return err
		}
	}
	switch strings.ToUpper(name) {
	case "TRACE":
		ll.level = LevelTrace
	case "DEBUG":
		ll.level = LevelDebug
	case "INFO":
		ll.level = LevelInfo
	case "NOTICE":
		ll.level = LevelNotice
	case "WARN", "WARNING":
		ll.level = LevelWarn
	case "ERROR":
		ll.level = LevelError
	case "EMERG", "EMERGENCY":
		ll.level = LevelEmergency
	default:
		return ErrInvalidLogLevel
	}

	ll.level += slog.Level(offset)
	ll.Level = strings.ToUpper(name)
	return nil
}

// UnmarshalJSON will intercept a JSON string to be converted to OutputFormat.
func (ll *LogLevel) UnmarshalJSON(data []byte) error {
	var level string
	if err := json.Unmarshal(data, &level); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from JSON: %w", err)
	}
	ll.Level = level

	return ll.DecodeLevel()
}

// UnmarshalYAML will intercept a YAML string to be converted to OutputFormat.
func (ll *LogLevel) UnmarshalYAML(data []byte) error {
	var level string
	if err := yaml.Unmarshal(data, &level); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from YAML: %w", err)
	}
	ll.Level = level

	return ll.DecodeLevel()
}

// UnmarshalTOML will intercept a TOML string to be converted to OutputFormat.
func (ll *LogLevel) UnmarshalTOML(data []byte) error {
	var level string
	if err := toml.Unmarshal(data, &level); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from TOML: %w", err)
	}
	ll.Level = level

	return ll.DecodeLevel()
}

// OutputFormat is the log record output formatting.
// Currently, JSON and TEXT are supported.
type OutputFormat string

// FromString will convert the given string to the matching OutputFormat.
func (of *OutputFormat) FromString(format string) error {
	switch strings.ToUpper(format) {
	case string(FormatText):
		*of = FormatText
	case string(FormatJSON):
		*of = FormatJSON
	default:
		return fmt.Errorf("invalid format: %s", format)
	}

	return nil
}

// UnmarshalJSON will intercept a JSON string to be converted to OutputFormat.
func (of *OutputFormat) UnmarshalJSON(data []byte) error {
	var format string
	if err := json.Unmarshal(data, &format); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from JSON: %w", err)
	}

	return of.FromString(format)
}

// UnmarshalYAML will intercept a YAML string to be converted to OutputFormat.
func (of *OutputFormat) UnmarshalYAML(data []byte) error {
	var format string
	if err := yaml.Unmarshal(data, &format); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from YAML: %w", err)
	}

	return of.FromString(format)
}

// UnmarshalTOML will intercept a TOML string to be converted to OutputFormat.
func (of *OutputFormat) UnmarshalTOML(data []byte) error {
	var format string
	if err := toml.Unmarshal(data, &format); err != nil {
		return fmt.Errorf("unmarshal OutputFormat from TOML: %w", err)
	}

	return of.FromString(format)
}

// Handler will return a slog.Handler that matches the configured output format (default=TEXT).
func (of *OutputFormat) Handler(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
	if *of == FormatJSON {
		return slog.NewJSONHandler(w, opts)
	}

	// Default to a TextHandler
	return slog.NewTextHandler(w, opts)
}

// Config is the root configuration for the logging library.
type Config struct {
	Console  ConsoleOutput
	File     FileOutput
	Handlers []slog.Handler
}

// Validate will check for common errors in the configuration.
func (c *Config) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [root]: %w", err)
		}
	}()
	// TODO: Validate recursively
	return nil
}

// OutputBase defines the common settings for an output type.
type OutputBase struct {
	// Disable this logging output.
	Disable bool
	// LogLevel handles the configuration of the current Log Level.
	LogLevel
	// Format of the log output, currently FormatText (default) and FormatJSON are supported.
	Format OutputFormat
	// IncludeSource will include the source code position of the log statement.
	IncludeSource bool
	// IncludeFullSource will include the directory for the source's filename.
	IncludeFullSource bool
}

func (OutputBase) JSONSchemaExtend(schema *jsonschema.Schema) {
	if schema == nil {
		schema = &jsonschema.Schema{}
	}

	if schema.Properties == nil {
		schema.Properties = jsonschema.NewProperties()
	}

	levelSchema, ok := schema.Properties.Get("Level")
	if !ok {
		levelSchema = &jsonschema.Schema{}
		schema.Properties.Set("Level", levelSchema)
	}

	levelSchema.Pattern = "^(?i)(trace|debug|info|notice|warning|warn|error|emerg|emergency)([+-][1-9][0-9]*)?$|^(\\d+)$"
}

// ConsoleOutput defines the settings specific to the console base output.
type ConsoleOutput struct {
	OutputBase
	// StdOut should only be enabled as a user preference, StdErr is designated for logging and non-interactive output.
	StdOut bool
}

// FileOutput defines the settings specific to the file based output.
type FileOutput struct {
	OutputBase
	// Path is the folder that logs should be written to. If not provided, file based logging will be disabled.
	Path        string `json:",omitempty" jsonschema:"title=Logging Path,example=./logs,default=Current Directory"`
	Rotate      Rotate
	loggingFile string
}

// JSONSchemaExtend extends the JSON schema for the FileOutput type.
// It adds conditions and requirements for the "Disable" and "Path" fields.
func (FileOutput) JSONSchemaExtend(schema *jsonschema.Schema) {
	// Add condition for the "Disable" field
	if schema.If == nil {
		schema.If = &jsonschema.Schema{}
	}

	if schema.If.Properties == nil {
		schema.If.Properties = jsonschema.NewProperties()
	}

	ifDisableSchema, ok := schema.If.Properties.Get("Disable")
	if !ok {
		ifDisableSchema = &jsonschema.Schema{}
		schema.If.Properties.Set("Disable", ifDisableSchema)
	}
	ifDisableSchema.Enum = []any{false}

	// Add requirement for the "Path" field if "Disable" is false
	if schema.Then == nil {
		schema.Then = &jsonschema.Schema{
			Required: []string{"Path"},
		}
	} else if !contains(schema.Then.Required, "Path") {
		schema.Then.Required = append(schema.Then.Required, "Path")
	}

	// Add condition and requirement for the "Path" field if "Disable" is true
	if schema.Else == nil {
		schema.Else = &jsonschema.Schema{
			If: &jsonschema.Schema{
				Not: &jsonschema.Schema{
					Required: []string{"Disable"},
				},
				Then: &jsonschema.Schema{
					Required: []string{"Path"},
				},
			},
		}
	} else if schema.Else.If == nil {
		schema.Else.If = &jsonschema.Schema{
			Not: &jsonschema.Schema{
				Required: []string{"Disable"},
			},
			Then: &jsonschema.Schema{
				Required: []string{"Path"},
			},
		}
	} else {
		if schema.Else.If.Not == nil {
			schema.Else.If.Not = &jsonschema.Schema{
				Required: []string{"Disable"},
			}
		} else if !contains(schema.Else.If.Not.Required, "Disable") {
			schema.Else.If.Not.Required = append(schema.Else.If.Not.Required, "Disable")
		}
		if schema.Else.If.Then == nil {
			schema.Else.If.Then = &jsonschema.Schema{
				Required: []string{"Path"},
			}
		} else if !contains(schema.Else.If.Then.Required, "Path") {
			schema.Else.If.Then.Required = append(schema.Else.If.Then.Required, "Path")
		}
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func (fo *FileOutput) GetPath() string {
	if fo.loggingFile != "" {
		return fo.loggingFile
	}

	name := filepath.Base(os.Args[0])
	fo.loggingFile = filepath.Join(fo.Path, name+".log")

	return fo.loggingFile
}

type Rotate struct {
	// Disable log rotation, enabled by default.
	Disable bool `json:",omitempty" jsonschema:"title=Disable Rotation,example=true,default=false"`
	// OnStart will trigger a log rotation on each start of the application.
	OnStart bool `json:",omitempty" jsonschema:"title=Rotate on Start,example=true,default=false"`
	// MaxSize (Megabytes) of a log file to trigger rotation.
	MaxSize int `json:",omitempty" jsonschema:"title=Max Filesize,example=1,default=5"`
	// Keep this many log files.
	Keep int `json:",omitempty" jsonschema:"title=Keep Files,example=3,default=4"`
	// MaxAge (in days) for a log file to triggering rotation.
	MaxAge int `json:",omitempty" jsonschema:"title=Max File Age,example=5,default=7"`
}
