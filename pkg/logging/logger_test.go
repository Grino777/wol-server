package logging

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"go.uber.org/zap/zapcore"
)

func TestNewLogger(t *testing.T) {
	t.Parallel()

	logger := NewLogger("", true)
	if logger == nil {
		t.Fatal("expected logger to never be nil")
	}
}

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	logger1 := DefaultLogger()
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	logger2 := DefaultLogger()
	if logger2 == nil {
		t.Fatal("expected logger to never be nil")
	}

	// Intentionally comparing identities here
	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
}

func TestContext(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	logger1 := FromContext(ctx)
	if logger1 == nil {
		t.Fatal("expected logger to never be nil")
	}

	ctx = WithLogger(ctx, logger1)

	logger2 := FromContext(ctx)
	if logger1 != logger2 {
		t.Errorf("expected %#v to be %#v", logger1, logger2)
	}
}

func TestLevelToZapLevls(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  zapcore.Level
	}{
		{input: LevelDebug, want: zapcore.DebugLevel},
		{input: LevelInfo, want: zapcore.InfoLevel},
		{input: LevelWarning, want: zapcore.WarnLevel},
		{input: LevelError, want: zapcore.ErrorLevel},
		{input: LevelCritical, want: zapcore.DPanicLevel},
		{input: LevelAlert, want: zapcore.PanicLevel},
		{input: LevelEmergency, want: zapcore.FatalLevel},
		{input: "unknown", want: zapcore.WarnLevel},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()

			got := levelToZapLevel(tc.input)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Fatalf("mismatch (-want, +got):\n%s", diff)
			}
		})
	}
}
