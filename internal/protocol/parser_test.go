package protocol

import (
	"errors"
	"testing"
)

func TestParseLine_Ping(t *testing.T) {
	cmd, err := ParseLine("PING")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Name != CmdPing {
		t.Errorf("expected command name %q, got %q", CmdPing, cmd.Name)
	}
	if len(cmd.Args) != 0 {
		t.Errorf("expected no arguments, got %d", len(cmd.Args))
	}
}

func TestParseLine_Set(t *testing.T) {
	cmd, err := ParseLine("SET mykey myvalue")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Name != CmdSet {
		t.Errorf("expected command name %q, got %q", CmdSet, cmd.Name)
	}
	if len(cmd.Args) != 2 {
		t.Errorf("expected 2 arguments, got %d", len(cmd.Args))
	}
	if cmd.Args[0] != "mykey" || cmd.Args[1] != "myvalue" {
		t.Errorf("expected arguments %q and %q, got %q and %q", "mykey", "myvalue", cmd.Args[0], cmd.Args[1])
	}
}

func TestParseLine_SetValueWithSpaces(t *testing.T) {
	cmd, err := ParseLine("SET msg hello godis")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cmd.Args[0] != "msg" {
		t.Fatalf("key = %q, want %q", cmd.Args[0], "msg")
	}
	if cmd.Args[1] != "hello godis" {
		t.Fatalf("value = %q, want %q", cmd.Args[1], "hello godis")
	}
}

func TestParseLine_Unknown(t *testing.T) {
	_, err := ParseLine("UNKNOWNCMD arg1 arg2")
	if !errors.Is(err, ErrUnknownCommand) {
		t.Fatalf("expected error %v, got %v", ErrUnknownCommand, err)
	}
}

func TestParseLine_Empty(t *testing.T) {
	_, err := ParseLine("   ")
	if !errors.Is(err, ErrEmptyCommand) {
		t.Fatalf("expected error %v, got %v", ErrEmptyCommand, err)
	}
}

func TestParseLine_InvalidArity(t *testing.T) {
	cases := []string{
		"PING now",
		"GET",
		"DEL",
		"EXISTS",
		"SET onlykey",
	}
	for _, line := range cases {
		_, err := ParseLine(line)
		if !errors.Is(err, ErrInvalidArity) {
			t.Fatalf("expected error %v for line %q, got %v", ErrInvalidArity, line, err)
		}
	}
}
