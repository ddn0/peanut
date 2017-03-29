package logwriter

import (
	"bytes"
	"log"
	"testing"
)

func TestSimple(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	lw := New(logger, nil)
	expected := "with newline\n"
	if _, err := lw.Write([]byte(expected)); err != nil {
		t.Fatalf("error writing: %s", err)
	}
	if err := lw.Flush(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if buf.String() != expected {
		t.Errorf("expected %q but found %q", expected, buf.String())
	}
}

func TestMissingNewline(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	lw := New(logger, nil)
	towrite := "without newline"
	expected := towrite + "\n"
	if _, err := lw.Write([]byte(towrite)); err != nil {
		t.Fatalf("error writing: %s", err)
	}
	if err := lw.Flush(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if buf.String() != expected {
		t.Errorf("expected %q but found %q", expected, buf.String())
	}
}

func TestMultipleWrites(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	lw := New(logger, nil)
	towrite := []string{"without newline", "newline\n", "without"}

	var expected bytes.Buffer
	for _, v := range towrite {
		expected.WriteString(v)
		if _, err := lw.Write([]byte(v)); err != nil {
			t.Fatalf("error writing: %s", err)
		}
	}
	expected.WriteString("\n")

	if err := lw.Flush(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if buf.String() != expected.String() {
		t.Errorf("expected %q but found %q", expected.String(), buf.String())
	}
}

func TestMultipleWritesMultipleNewlines(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)
	lw := New(logger, nil)
	towrite := []string{"without newline\n\n", "new\nline\n", "without"}

	var expected bytes.Buffer
	for _, v := range towrite {
		expected.WriteString(v)
		if _, err := lw.Write([]byte(v)); err != nil {
			t.Fatalf("error writing: %s", err)
		}
	}
	expected.WriteString("\n")

	if err := lw.Flush(); err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	if buf.String() != expected.String() {
		t.Errorf("expected %q but found %q", expected.String(), buf.String())
	}
}
