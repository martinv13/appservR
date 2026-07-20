package config

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

// captureStdout redirects os.Stdout for the duration of fn and returns
// whatever was written to it.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("failed to read captured output: %v", err)
	}
	return buf.String()
}

func TestLoggerDebugOnlyPrintsAtDebugLevel(t *testing.T) {
	debug := NewLogger(logLevels.DEBUG)
	out := captureStdout(t, func() { debug.Debug("hello debug") })
	if !strings.Contains(out, "hello debug") {
		t.Errorf("expected Debug level logger to print Debug messages, got %q", out)
	}

	info := NewLogger(logLevels.INFO)
	out = captureStdout(t, func() { info.Debug("hidden debug") })
	if strings.Contains(out, "hidden debug") {
		t.Errorf("expected Info level logger to suppress Debug messages, got %q", out)
	}
}

func TestLoggerInfoPrintsAtDebugAndInfoLevels(t *testing.T) {
	for _, level := range []int{logLevels.DEBUG, logLevels.INFO} {
		l := NewLogger(level)
		out := captureStdout(t, func() { l.Info("hello info") })
		if !strings.Contains(out, "hello info") {
			t.Errorf("expected level %d logger to print Info messages, got %q", level, out)
		}
	}

	warning := NewLogger(logLevels.WARNING)
	out := captureStdout(t, func() { warning.Info("hidden info") })
	if strings.Contains(out, "hidden info") {
		t.Errorf("expected Warning level logger to suppress Info messages, got %q", out)
	}
}

func TestLoggerWarningSuppressedAtErrorLevel(t *testing.T) {
	warning := NewLogger(logLevels.WARNING)
	out := captureStdout(t, func() { warning.Warning("hello warning") })
	if !strings.Contains(out, "hello warning") {
		t.Errorf("expected Warning level logger to print Warning messages, got %q", out)
	}

	errLevel := NewLogger(logLevels.ERROR)
	out = captureStdout(t, func() { errLevel.Warning("hidden warning") })
	if strings.Contains(out, "hidden warning") {
		t.Errorf("expected Error level logger to suppress Warning messages, got %q", out)
	}
}

func TestLoggerErrorAlwaysPrints(t *testing.T) {
	for _, level := range []int{logLevels.DEBUG, logLevels.INFO, logLevels.WARNING, logLevels.ERROR} {
		l := NewLogger(level)
		out := captureStdout(t, func() { l.Error("boom") })
		if !strings.Contains(out, "boom") {
			t.Errorf("expected level %d logger to always print Error messages, got %q", level, out)
		}
	}
}
