package nmcslog

import (
	"fmt"
	"log/slog"
	"os"
)

var (
	defaultLogger *slog.Logger = slog.Default()
	debugLogger   *slog.Logger = nil
)

// Logger will return the default Slog logger which can be updated via the SetDefaultLogger function.
func Logger() *slog.Logger {
	return defaultLogger
}

// SetDefaultLogger will allow the user to specify a custom logger to be used as the default logger.
func SetDefaultLogger(l *slog.Logger) *slog.Logger {
	defaultLogger = l
	return Logger()
}

// DebugLogger will return a default Slog logger at level DEBUG which can be updated via the SetDebugLogger function.
func DebugLogger() *slog.Logger {
	if debugLogger == nil {
		debugLogger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			AddSource:   true,
			Level:       slog.LevelDebug,
			ReplaceAttr: nil,
		}))
	}
	return defaultLogger
}

// SetDebugLogger will allow the user to specify a custom logger to be used as the default DEBUG logger.
func SetDebugLogger(l *slog.Logger) *slog.Logger {
	defaultLogger = l
	return Logger()
}

// GetConfiguredLogger will return a logger that is configured according to the given configuration.
func GetConfiguredLogger(logConfig *Config) (logger *slog.Logger, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("nmcslog: validate config [configure logger]: %w", err)
		}
	}()

	if err = logConfig.Validate(); err != nil {
		return nil, fmt.Errorf("invalid logger configuration: %w", err)
	}

	logger, err = logConfig.GetHandlers()
	if err != nil {
		return nil, fmt.Errorf("get configured logger: %w", err)
	}

	return logger, nil
}
