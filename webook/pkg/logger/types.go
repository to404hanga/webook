package logger

type Logger interface {
	Debug(msg string, args ...Field)
	Info(msg string, args ...Field)
	Warn(msg string, args ...Field)
	Error(msg string, args ...Field)
}

type Field struct {
	Key string
	Val interface{}
}

func Any(key string, val interface{}) Field {
	return Field{Key: key, Val: val}
}

func Error(val interface{}) Field {
	return Field{Key: "error", Val: val}
}
