package nmcslog

import (
	"log/slog"
	"path/filepath"
)

type AttributeFunc func([]string, slog.Attr) slog.Attr

// WrapAttributeFuncs will wrap multiple Attribute Middlewares around a given Attribute Function.
func WrapAttributeFuncs(af ...AttributeFunc) AttributeFunc {
	if len(af) < 1 {
		// return func(_ []string, a slog.Attr) slog.Attr { return a}
		return nil
	}

	return func(groups []string, a slog.Attr) slog.Attr {
		for _, fn := range af {
			a = fn(groups, a)
		}
		return a
	}
}

func AttrRemoveFullSource(groups []string, a slog.Attr) slog.Attr {
	// Remove the directory from the source's filename.
	if a.Key == slog.SourceKey {
		a.Value = slog.StringValue(filepath.Base(a.Value.String()))
	}

	return a
}

func AttrFixCustomLogLevelNames(groups []string, a slog.Attr) slog.Attr {
	// Properly name log levels in output such as NOTICE rather than INFO+2
	if a.Key == slog.LevelKey {
		level := a.Value.Any().(slog.Level)
		levelLabel, exists := CustomLevelNames[level]
		if !exists {
			levelLabel = level.String()
		}

		a.Value = slog.StringValue(levelLabel)
	}
	return a
}
