package nmcslog

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/invopop/jsonschema"
	"github.com/invopop/yaml"
)

// OutputHandler defines the common settings for an output type.
type OutputHandler struct {
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
	// Middleware is an array of middleware funcs to modify the log record prior to calling the handler.
	// https://github.com/samber/slog-multi#custom-middleware
	Middleware []MiddlewareFunc
	// AttributeFuncs is an array of functions to modify log record attributes.
	AttributeFuncs []AttributeFunc
}

func (ob *OutputHandler) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [output base]: %w", err)
		}
	}()

	if ob.Disable {
		return nil
	}
	if err = ob.LogLevel.Validate(); err != nil {
		return err
	}
	if err = ob.Format.Validate(); err != nil {
		return err
	}
	return nil
}

func (OutputHandler) JSONSchemaExtend(schema *jsonschema.Schema) {
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

	levelSchema.Pattern = "^(?i)(trace|debug|info|notice|warning|warn|error|fatal)([+-][1-9][0-9]*)?$|^(\\d+)$"
}

func (ob *OutputHandler) GetHandler(w io.Writer) (handler slog.Handler, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: get handler: %w", err)
		}
	}()

	if ob.Disable {
		return nil, fmt.Errorf("[%s] %w", ob.Format, ErrHandlerDisabled)
	}

	attrFuncs := []AttributeFunc{
		AttrFixCustomLogLevelNames,
	}

	if !ob.IncludeFullSource {
		attrFuncs = append(attrFuncs, AttrRemoveFullSource)
	}

	if len(ob.AttributeFuncs) > 0 {
		attrFuncs = append(attrFuncs, ob.AttributeFuncs...)
	}

	handlerOpts := &slog.HandlerOptions{
		AddSource:   ob.IncludeSource,
		Level:       ob.LogLevel.level,
		ReplaceAttr: WrapAttributeFuncs(attrFuncs...),
	}

	logHandler := ob.Format.Handler(w, handlerOpts)

	if len(ob.Middleware) > 0 {
		return WrapMiddlewareFuncs(logHandler, ob.Middleware...), err
	}

	return logHandler, nil
}

// LogLevel handles the decoding and parsing of a given log level to the slog log level.
type LogLevel struct {
	// Level to cutoff log messages, anything below this level will be dropped.
	Level string
	// Level is the converted Level from either an integer or string.
	level   *slog.LevelVar
	decoded bool
}

func (ll *LogLevel) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [log level]: %w", err)
		}
	}()

	if !ll.decoded {
		if err = ll.DecodeLevel(); err != nil {
			return fmt.Errorf("decoding log level: %w", err)
		}
	}

	return nil
}

func (ll *LogLevel) DecodeLevel() (err error) {
	if ll.decoded {
		return nil
	}
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: decode level string %q: %w", ll.Level, err)
		}
	}()

	if ll.level == nil {
		ll.level = &slog.LevelVar{}
	}

	if ll.Level == "" {
		ll.level.Set(slog.LevelInfo)
		return nil
	}
	if l, err := strconv.ParseInt(ll.Level, 10, 64); err == nil {
		ll.level.Set(slog.Level(l))
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
		ll.level.Set(LevelTrace)
	case "DEBUG":
		ll.level.Set(LevelDebug)
	case "INFO":
		ll.level.Set(LevelInfo)
	case "NOTICE":
		ll.level.Set(LevelNotice)
	case "WARN", "WARNING":
		ll.level.Set(LevelWarn)
	case "ERROR":
		ll.level.Set(LevelError)
	case "FATAL":
		ll.level.Set(LevelFatal)
	default:
		return ErrInvalidLogLevel
	}

	ll.level.Set(ll.level.Level() + slog.Level(offset))
	ll.Level = strings.ToUpper(name)
	ll.decoded = true
	return nil
}

func (ll *LogLevel) GetSlogLevel() (levelVar slog.Level, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: get log level %q: %w", ll.Level, err)
		}
	}()

	if !ll.decoded {
		if err = ll.DecodeLevel(); err != nil {
			return 0, fmt.Errorf("get log level: %w", err)
		}
	}

	return ll.level.Level(), nil
}

func (ll *LogLevel) SetSlogLevel(newLevel string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: set log level %q: %w", ll.Level, err)
		}
	}()

	oldLevel := ll.Level
	ll.Level = newLevel
	if err = ll.DecodeLevel(); err != nil {
		ll.Level = oldLevel
		return fmt.Errorf("decoding new level: %w", err)
	}

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

func (of *OutputFormat) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [output format]: %w", err)
		}
	}()
	if err = of.FromString(string(*of)); err != nil {
		return err
	}
	return nil
}

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
