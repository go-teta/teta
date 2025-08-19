package teta

import (
	"io"
	"log/slog"
)

func newLogger(w io.Writer) *TetaLogger {
	logger := setupLogger(w)

	slog.SetDefault(logger)
	return &TetaLogger{
		handler: logger,
		output:  w,
	}
}

type TetaLogger struct {
	handler *slog.Logger
	output  io.Writer
}

type Logger interface {
	Output() io.Writer
	SetOutput(w io.Writer)
	WithFields(args ...any) Logger

	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
}

func (l *TetaLogger) Output() io.Writer {
	return l.output
}

func (l *TetaLogger) SetOutput(w io.Writer) {
	l.output = w
	l.handler = setupLogger(w)
}

func (l *TetaLogger) WithFields(args ...any) Logger {
	return &TetaLogger{
		handler: l.handler.With(args...),
		output:  l.output,
	}
}

func (l *TetaLogger) Debug(msg string, args ...any) {
	l.handler.Debug(msg, args...)
}

func (l *TetaLogger) Info(msg string, args ...any) {
	l.handler.Info(msg, args...)
}

func (l *TetaLogger) Warn(msg string, args ...any) {
	l.handler.Warn(msg, args...)
}

func (l *TetaLogger) Error(msg string, args ...any) {
	l.handler.Error(msg, args...)
}

func (l *TetaLogger) Fatal(msg string, args ...any) {
	l.handler.Error(msg, args...)
	panic("FATAL: " + msg)
}

func setupLogger(w io.Writer) *slog.Logger {
	jsonHandler := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: slog.LevelDebug,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			switch a.Key {
			case slog.TimeKey:
				a.Key = "timestamp"
			case slog.LevelKey:
				a.Key = "level"
			case slog.MessageKey:
				a.Key = "message"
			}
			return a
		},
	})

	return slog.New(jsonHandler)
}
