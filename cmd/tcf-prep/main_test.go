package main

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHelp(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"--help"}, &out, &errBuf); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	for _, want := range []string{"Usage:", "scrape", "quiz", "run", "version"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("help output missing %q:\n%s", want, out.String())
		}
	}
}

func TestRunNoArgsPrintsUsage(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run(nil, &out, &errBuf); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("output missing usage:\n%s", out.String())
	}
}

func TestRunVersion(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"version"}, &out, &errBuf); code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}
	if !strings.Contains(out.String(), version) {
		t.Errorf("output = %q, want to contain %q", out.String(), version)
	}
}

func TestRunUnknownCommand(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"frobnicate"}, &out, &errBuf); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if !strings.Contains(errBuf.String(), "unknown command") {
		t.Errorf("stderr = %q, want to contain %q", errBuf.String(), "unknown command")
	}
}

func TestRunQuizRequiresGeminiKey(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "")

	var out, errBuf bytes.Buffer
	if code := run([]string{"quiz"}, &out, &errBuf); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "GEMINI_API_KEY") {
		t.Errorf("stderr = %q, want to mention GEMINI_API_KEY", errBuf.String())
	}
}

func TestRunQuizRejectsInvalidMode(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-key")

	var out, errBuf bytes.Buffer
	if code := run([]string{"quiz", "--mode", "xx"}, &out, &errBuf); code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
}

func TestRunTestScrapeAlias(t *testing.T) {
	var out, errBuf bytes.Buffer
	if code := run([]string{"--test-scrape"}, &out, &errBuf); code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if !strings.Contains(errBuf.String(), "not implemented") {
		t.Errorf("stderr = %q, want to contain %q", errBuf.String(), "not implemented")
	}
}
