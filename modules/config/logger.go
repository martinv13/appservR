package config

import (
	"fmt"
)

type Logger struct {
	level int
}

var logLevels = struct {
	DEBUG   int
	INFO    int
	WARNING int
	ERROR   int
}{
	DEBUG:   0,
	INFO:    1,
	WARNING: 2,
	ERROR:   3,
}

func NewLogger(level int) Logger {
	return Logger{
		level: level,
	}
}

func (l *Logger) Debug(s string) {
	if l.level == 0 {
		fmt.Println(s)
	}
}

func (l *Logger) Info(s string) {
	if l.level <= 1 {
		fmt.Println(s)
	}
}

func (l *Logger) Warning(s string) {
	if l.level <= 2 {
		fmt.Println(s)
	}
}

func (l *Logger) Error(s string) {
	if l.level <= 3 {
		fmt.Println(s)
	}
}
