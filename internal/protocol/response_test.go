package protocol

import (
	"testing"
)

func TestSimpleString(t *testing.T) {
	result := SimpleString("OK")
	expected := "+OK\r\n"
	if result != expected {
		t.Errorf("SimpleString(\"OK\") = %q; want %q", result, expected)
	}
}

func TestError(t *testing.T) {
	got := Error("bad command")
	want := "-ERR bad command\r\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestIntegerBool(t *testing.T) {
	if IntegerBool(true) != ":1\r\n" {
		t.Errorf("IntegerBool(true) = %q; want %q", IntegerBool(true), ":1\r\n")
	}
	if IntegerBool(false) != ":0\r\n" {
		t.Errorf("IntegerBool(false) = %q; want %q", IntegerBool(false), ":0\r\n")
	}
}

func TestBulkString(t *testing.T) {
	result := BulkString("hello")
	expected := "$5\r\nhello\r\n"
	if result != expected {
		t.Errorf("BulkString(\"hello\") = %q; want %q", result, expected)
	}
}

func TestNilBulkString(t *testing.T) {
	result := NilBulkString()
	expected := "$-1\r\n"
	if result != expected {
		t.Fatalf("NilBulkString() = %q; want %q", result, expected)
	}
}

func TestConveninceResponses(t *testing.T) {
	if Pong() != "+PONG\r\n" {
		t.Fatalf("Pong() = %q; want %q", Pong(), "+PONG\r\n")
	}
	if OK() != "+OK\r\n" {
		t.Fatalf("OK() = %q; want %q", OK(), "+OK\r\n")
	}
}
