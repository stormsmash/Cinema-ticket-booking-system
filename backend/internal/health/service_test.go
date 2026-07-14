package health

import (
	"context"
	"errors"
	"testing"
)

func TestServiceReportsReady(t *testing.T) {
	service := NewService(map[string]CheckFunc{
		"mongodb": func(context.Context) error { return nil },
		"redis":   func(context.Context) error { return nil },
	})

	report := service.Check(context.Background())

	if report.Status != StatusReady {
		t.Fatalf("expected status %q, got %q", StatusReady, report.Status)
	}
	if report.Checks["mongodb"] != CheckOK || report.Checks["redis"] != CheckOK {
		t.Fatalf("expected all dependency checks to pass, got %#v", report.Checks)
	}
}

func TestServiceReportsUnavailableDependency(t *testing.T) {
	service := NewService(map[string]CheckFunc{
		"mongodb": func(context.Context) error { return nil },
		"redis":   func(context.Context) error { return errors.New("redis unavailable") },
	})

	report := service.Check(context.Background())

	if report.Status != StatusNotReady {
		t.Fatalf("expected status %q, got %q", StatusNotReady, report.Status)
	}
	if report.Checks["redis"] != CheckFailed {
		t.Fatalf("expected Redis check to fail, got %#v", report.Checks)
	}
}
