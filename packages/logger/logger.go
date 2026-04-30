package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	zl zerolog.Logger
}

func New(service string) *Logger {
	zerolog.TimeFieldFormat = time.RFC3339Nano

	level := parseLevel(os.Getenv("LOG_LEVEL"))
	zerolog.SetGlobalLevel(level)

	zl := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Str("service", service).
		Logger()

	return &Logger{zl: zl}
}

func parseLevel(s string) zerolog.Level {
	switch s {
	case "trace":
		return zerolog.TraceLevel
	case "debug":
		return zerolog.DebugLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	default:
		return zerolog.InfoLevel
	}
}

func (l *Logger) Trace(msg string, fields ...Field) {
	e := l.zl.Trace()
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) Debug(msg string, fields ...Field) {
	e := l.zl.Debug()
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) Info(msg string, fields ...Field) {
	e := l.zl.Info()
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) Warn(msg string, fields ...Field) {
	e := l.zl.Warn()
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) Error(msg string, err error, fields ...Field) {
	e := l.zl.Error().Err(err)
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) Fatal(msg string, err error, fields ...Field) {
	e := l.zl.Fatal().Err(err)
	applyFields(e, fields).Msg(msg)
}

func (l *Logger) With(fields ...Field) *Logger {
	ctx := l.zl.With()
	for _, f := range fields {
		ctx = f.applyToContext(ctx)
	}
	return &Logger{zl: ctx.Logger()}
}
