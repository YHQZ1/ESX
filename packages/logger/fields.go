package logger

import "github.com/rs/zerolog"

type Field struct {
	key   string
	value any
}

func Str(key, val string) Field         { return Field{key, val} }
func Int(key string, val int) Field     { return Field{key, val} }
func Int64(key string, val int64) Field { return Field{key, val} }
func Bool(key string, val bool) Field   { return Field{key, val} }
func Any(key string, val any) Field     { return Field{key, val} }

func (f Field) applyToContext(ctx zerolog.Context) zerolog.Context {
	switch v := f.value.(type) {
	case string:
		return ctx.Str(f.key, v)
	case int:
		return ctx.Int(f.key, v)
	case int64:
		return ctx.Int64(f.key, v)
	case bool:
		return ctx.Bool(f.key, v)
	default:
		return ctx.Interface(f.key, v)
	}
}

func applyFields(e *zerolog.Event, fields []Field) *zerolog.Event {
	for _, f := range fields {
		switch v := f.value.(type) {
		case string:
			e = e.Str(f.key, v)
		case int:
			e = e.Int(f.key, v)
		case int64:
			e = e.Int64(f.key, v)
		case bool:
			e = e.Bool(f.key, v)
		default:
			e = e.Interface(f.key, v)
		}
	}
	return e
}
