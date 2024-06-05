package nmcslog

import (
	"context"
	"log/slog"

	"github.com/fatih/color"
	slogmulti "github.com/samber/slog-multi"
)

type ctxKey string

type MiddlewareFunc func(slog.Handler) slog.Handler

const (
	slogFields ctxKey = "slog_fields"
)

// WrapMiddlewareFuncs will wrap multiple Attribute Middlewares around a given Attribute Function.
func WrapMiddlewareFuncs(handler slog.Handler, mwf ...MiddlewareFunc) slog.Handler {
	if len(mwf) < 1 {
		// return func(_ []string, a slog.Attr) slog.Attr { return a}
		return handler
	}

	wrapped := handler

	// loop in reverse to preserve middleware order
	for i := len(mwf) - 1; i >= 0; i-- {
		wrapped = mwf[i](wrapped)
	}

	return wrapped
}

// MWHandleColors (WIP) is a middleware that will colorize the LEVEL text.
type MWHandleColors struct {
	ColorMap     map[slog.Level]color.Attribute
	DefaultColor color.Attribute
}

func (hc *MWHandleColors) SetLevelColor(level slog.Level, format color.Attribute) {
	if hc.ColorMap == nil {
		hc.ColorMap = map[slog.Level]color.Attribute{level: format}
	} else {
		hc.ColorMap[level] = format
	}
}

func (hc *MWHandleColors) RemoveLevelColor(level slog.Level) {
	if hc.ColorMap != nil {
		delete(hc.ColorMap, level)
	}
}

func (hc *MWHandleColors) SetDefaultColors() {
	hc.SetLevelColor(LevelTrace, color.FgWhite)
	hc.SetLevelColor(LevelDebug, color.FgCyan)
	hc.SetLevelColor(LevelInfo, color.FgGreen)
	hc.SetLevelColor(LevelNotice, color.FgBlue)
	hc.SetLevelColor(LevelWarn, color.FgMagenta)
	hc.SetLevelColor(LevelError, color.FgYellow)
	hc.SetLevelColor(LevelFatal, color.FgRed)
}

func (hc *MWHandleColors) Middleware() slogmulti.Middleware {
	return slogmulti.NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			if hc.ColorMap == nil {
				// return fmt.Errorf("middleware color map undefined")
				return next(ctx, record)
			}

			level := record.Level.String()
			format, exists := hc.ColorMap[record.Level]
			if !exists {
				return next(ctx, record)
			}
			level = color.New(format).Sprint(level)
			// TODO: How to actually print the level?

			return next(ctx, record)
		})
}

// MWHandleContext is a middleware that will extract slog.Attr attributes from the context.
// https://betterstack.com/community/guides/logging/logging-in-go/#using-the-context-package-with-slog
type MWHandleContext struct{}

func (hc *MWHandleContext) Middleware() slogmulti.Middleware {
	return slogmulti.NewHandleInlineMiddleware(
		func(ctx context.Context, record slog.Record, next func(context.Context, slog.Record) error) error {
			if attrs, ok := ctx.Value(slogFields).([]slog.Attr); ok {
				for _, v := range attrs {
					record.AddAttrs(v)
				}
			}
			return next(ctx, record)
		})
}

// AppendCtx adds a slog attribute to the provided context so that it will be
// included in any Record created with such context
func AppendCtx(parent context.Context, attr slog.Attr) context.Context {
	if parent == nil {
		parent = context.Background()
	}

	if v, ok := parent.Value(slogFields).([]slog.Attr); ok {
		return context.WithValue(parent, slogFields, append(v, attr))
	}

	return context.WithValue(parent, slogFields, []slog.Attr{attr})
}
