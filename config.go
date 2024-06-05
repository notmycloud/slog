package nmcslog

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/invopop/jsonschema"
	slogmulti "github.com/samber/slog-multi"
	"gopkg.in/natefinch/lumberjack.v2"
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
	// LevelFatal defines the Emergency Log Level (12)
	LevelFatal = slog.LevelError + 4

	// DefaultRotateSize is the default max log filesize in Megabytes.
	DefaultRotateSize = 5
	// DefaultRotateKeep is the default number of log files to keep.
	DefaultRotateKeep = 4
	// DefaultRotateAge is the default max age of a log file in days.
	DefaultRotateAge = 7
)

var (
	ErrInvalidLogLevel  = errors.New("invalid log level")
	ErrHandlerDisabled  = errors.New("handler disabled")
	ErrRotatorDisabled  = errors.New("rotator disabled")
	ErrNoHandersEnabled = errors.New("no handlers are enabled")
)

// CustomLevelNames will replace the SLOG Level output with the given names rather than the lower level + increment.
var CustomLevelNames = map[slog.Leveler]string{
	LevelTrace:  "TRACE",
	LevelNotice: "NOTICE",
	LevelFatal:  "FATAL",
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

	if err = c.Console.Validate(); err != nil {
		return err
	}
	if err = c.File.Validate(); err != nil {
		return err
	}

	return nil
}

func (c *Config) GetHandlers() (logger *slog.Logger, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: creating logger [root]: %w", err)
		}
	}()

	logHandlers := c.Handlers
	if !c.Console.Disable {
		var handler slog.Handler
		handler, err = c.Console.GetHandler()
		if err != nil {
			return nil, fmt.Errorf("getting console log handler: %w", err)
		}
		logHandlers = append(logHandlers, handler)
	}

	if !c.File.Disable {
		var handler slog.Handler
		handler, err = c.File.GetHandler()
		if err != nil {
			return nil, fmt.Errorf("getting file log handler: %w", err)
		}
		logHandlers = append(logHandlers, handler)
	}

	if len(logHandlers) == 0 {
		return nil, ErrNoHandersEnabled
	}

	var logHandler slog.Handler

	if len(logHandlers) == 1 {
		logHandler = logHandlers[0]
	} else {
		logHandler = slogmulti.Fanout(logHandlers...)
	}

	return slog.New(logHandler), nil
}

// ConsoleOutput defines the settings specific to the console base output.
type ConsoleOutput struct {
	OutputHandler
	// StdOut should only be enabled as a user preference, StdErr is designated for logging and non-interactive output.
	StdOut bool
}

func (co *ConsoleOutput) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [console output]: %w", err)
		}
	}()

	return co.OutputHandler.Validate()
}

func (co *ConsoleOutput) GetHandler() (handler slog.Handler, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: get handler [console output]: %w", err)
		}
	}()

	if co.Disable {
		return nil, fmt.Errorf("[%s] %w", co.Format, ErrHandlerDisabled)
	}

	handler, err = co.OutputHandler.GetHandler(os.Stderr)
	if err != nil {
		return nil, fmt.Errorf("getting handler [console]: %w", err)
	}

	return handler, nil
}

// FileOutput defines the settings specific to the file based output.
type FileOutput struct {
	OutputHandler
	// Path is the folder that logs should be written to. If not provided, file based logging will be disabled.
	Path        string `json:",omitempty" jsonschema:"title=Logging Path,example=./logs,default=Current Directory"`
	Filename    string
	Rotate      Rotate
	loggingFile string
}

func (fo *FileOutput) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [file output]: %w", err)
		}
	}()

	if err = fo.OutputHandler.Validate(); err != nil {
		return err
	}
	if err = validatePath(fo.GetPath()); err != nil {
		return fmt.Errorf("invalid logging path [%s]: %w", fo.loggingFile, err)
	}
	if err = fo.Rotate.Validate(); err != nil {
		return err
	}
	return nil
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

func (fo *FileOutput) GetPath() string {
	if fo.loggingFile != "" {
		return fo.loggingFile
	}

	if fo.Filename == "" {
		fo.Filename = filepath.Base(os.Args[0])
	}
	fo.loggingFile = filepath.Join(fo.Path, fo.Filename+".log")

	return fo.loggingFile
}

func (fo *FileOutput) GetHandler() (handler slog.Handler, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: get handler [file output]: %w", err)
		}
	}()

	if fo.Disable {
		return nil, fmt.Errorf("[%s] %w", fo.Format, ErrHandlerDisabled)
	}

	var output io.Writer
	if !fo.Rotate.Disable {
		if output, err = fo.Rotate.GetWriter(fo.GetPath()); err != nil {
			return nil, fmt.Errorf("getting log rotator: %w", err)
		}
	} else {
		logFile, err := os.OpenFile(
			fo.GetPath(),
			os.O_APPEND|os.O_CREATE|os.O_WRONLY,
			0o644, //nolint:gomnd
		)
		if err != nil {
			return nil, fmt.Errorf("create/open log file [%s]: %w", fo.GetPath(), err)
		}

		output = logFile
	}

	handler, err = fo.OutputHandler.GetHandler(output)
	if err != nil {
		return nil, fmt.Errorf("getting handler [file]: %w", err)
	}

	return handler, nil
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

func (r *Rotate) Validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [rotate]: %w", err)
		}
	}()
	if r.Disable {
		return nil
	}

	if r.MaxSize < 1 {
		return fmt.Errorf("invalid MaxSize [%d]", r.MaxSize)
	}
	if r.Keep < 1 {
		return fmt.Errorf("invalid Keep [%d]", r.Keep)
	}
	if r.MaxAge < 1 {
		return fmt.Errorf("invalid MaxAge [%d]", r.MaxAge)
	}
	return nil
}

func (r *Rotate) GetWriter(path string) (w io.Writer, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: get file rotator: %w", err)
		}
	}()

	if r.Disable {
		return nil, fmt.Errorf("log rotation disabled: %w", ErrRotatorDisabled)
	}

	rotator := &lumberjack.Logger{
		Filename:   path,
		MaxSize:    r.MaxSize, // megabytes
		MaxBackups: r.Keep,
		MaxAge:     r.MaxAge, // days
	}

	if r.OnStart {
		if err := rotator.Rotate(); err != nil {
			return nil, fmt.Errorf("rotating logs on startup: %w", err)
		}
	}

	return rotator, nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
