package example

import (
	"context"
	"log/slog"

	nmcslog "github.com/notmycloud/slog"
)

func Example_getConfiguredLogger() {
	logLevel := nmcslog.LogLevel{Level: nmcslog.LevelNotice.String()}
	baseConfig := nmcslog.OutputBase{
		Disable:           false,
		LogLevel:          logLevel,
		Format:            nmcslog.FormatText,
		IncludeSource:     true,
		IncludeFullSource: true,
	}
	consoleConfig := nmcslog.ConsoleOutput{
		OutputBase: baseConfig,
		StdOut:     false,
	}
	rotateConfig := nmcslog.Rotate{
		Disable: false,
		OnStart: true,
		MaxSize: 1,
		Keep:    5,
		MaxAge:  7,
	}
	fileConfig := nmcslog.FileOutput{
		OutputBase: baseConfig,
		Path:       "/tmp",
		Rotate:     rotateConfig,
	}
	customHandlerConfig := nmcslog.FileOutput{
		OutputBase: nmcslog.OutputBase{},
		Path:       "/tmp",
		Filename:   "custom",
		Rotate:     rotateConfig,
	}
	customHandler, err := customHandlerConfig.GetHandler()
	if err != nil {
		panic(err)
	}

	config := nmcslog.Config{
		Console: consoleConfig,
		File:    fileConfig,
		Handlers: []slog.Handler{
			customHandler,
		},
	}

	if err := config.Validate(); err != nil {
		panic(err)
	}

	logger, err := nmcslog.GetConfiguredLogger(&config)
	if err != nil {
		panic(err)
	}

	logger.Info("INITIAL LOGGER!")
	namedLogger := logger.With(slog.String("name", "Example_getConfiguredLogger"))
	namedLogger.Log(context.Background(), nmcslog.LevelTrace, "TRACE MESSAGE")
	namedLogger.Debug("DEBUG MESSAGE")
	namedLogger.Info("INFO MESSAGE")
	namedLogger.Log(context.Background(), nmcslog.LevelNotice, "NOTICE MESSAGE")
	namedLogger.Warn("WARNING MESSAGE")
	namedLogger.Error("ERROR MESSAGE")
	namedLogger.Log(context.Background(), nmcslog.LevelEmergency, "EMERGENCY MESSAGE")
	print("LOGGER COMPLETE")
	// NOTE: print() does not count as output!
	// NOTE: The // Output: comment must be "alone"

	// Output:
}
