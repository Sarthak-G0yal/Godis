package server

import (
	"errors"
	"testing"

	"mini-redis/internal/protocol"
	"mini-redis/internal/storage"
)

type mockAppender struct {
	commands []protocol.Command
	err      error
}

func (m *mockAppender) Append(cmd protocol.Command) error {
	m.commands = append(m.commands, cmd)
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestExecutorPing(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	got := exec.Execute(protocol.Command{Name: protocol.CmdPing})
	want := "+PONG\r\n"
	if got != want {
		t.Fatalf("PING response = %q, want %q", got, want)
	}
}

func TestExecutorSetAndGetHit(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	setResp := exec.Execute(protocol.Command{Name: protocol.CmdSet, Args: []string{"name", "mini-redis"}})
	if setResp != "+OK\r\n" {
		t.Fatalf("SET response = %q, want %q", setResp, "+OK\\r\\n")
	}

	getResp := exec.Execute(protocol.Command{Name: protocol.CmdGet, Args: []string{"name"}})
	if getResp != "$10\r\nmini-redis\r\n" {
		t.Fatalf("GET hit response = %q, want %q", getResp, "$10\\r\\nmini-redis\\r\\n")
	}
}

func TestExecutorGetMiss(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	got := exec.Execute(protocol.Command{Name: protocol.CmdGet, Args: []string{"missing"}})
	want := "$-1\r\n"
	if got != want {
		t.Fatalf("GET miss response = %q, want %q", got, want)
	}
}

func TestExecutorDelHitAndMiss(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	_ = exec.Execute(protocol.Command{Name: protocol.CmdSet, Args: []string{"k1", "v1"}})

	delHit := exec.Execute(protocol.Command{Name: protocol.CmdDel, Args: []string{"k1"}})
	if delHit != ":1\r\n" {
		t.Fatalf("DEL hit response = %q, want %q", delHit, ":1\\r\\n")
	}

	delMiss := exec.Execute(protocol.Command{Name: protocol.CmdDel, Args: []string{"k1"}})
	if delMiss != ":0\r\n" {
		t.Fatalf("DEL miss response = %q, want %q", delMiss, ":0\\r\\n")
	}
}

func TestExecutorExists(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	_ = exec.Execute(protocol.Command{Name: protocol.CmdSet, Args: []string{"feature", "on"}})

	existsTrue := exec.Execute(protocol.Command{Name: protocol.CmdExists, Args: []string{"feature"}})
	if existsTrue != ":1\r\n" {
		t.Fatalf("EXISTS true response = %q, want %q", existsTrue, ":1\\r\\n")
	}

	existsFalse := exec.Execute(protocol.Command{Name: protocol.CmdExists, Args: []string{"unknown"}})
	if existsFalse != ":0\r\n" {
		t.Fatalf("EXISTS false response = %q, want %q", existsFalse, ":0\\r\\n")
	}
}

func TestExecutorBadArity(t *testing.T) {
	exec := NewExecutor(storage.NewMemoryStorage(), nil)

	cases := []protocol.Command{
		{Name: protocol.CmdPing, Args: []string{"extra"}},
		{Name: protocol.CmdSet, Args: []string{"only-key"}},
		{Name: protocol.CmdGet, Args: nil},
		{Name: protocol.CmdDel, Args: []string{}},
		{Name: protocol.CmdExists, Args: []string{"k1", "extra"}},
	}

	for _, cmd := range cases {
		got := exec.Execute(cmd)
		if len(got) == 0 || got[:5] != "-ERR " {
			t.Fatalf("expected error response for %#v, got %q", cmd, got)
		}
	}
}

func TestExecutorAppendHookOnlySuccessfulWrites(t *testing.T) {
	appender := &mockAppender{}
	exec := NewExecutor(storage.NewMemoryStorage(), appender)

	_ = exec.Execute(protocol.Command{Name: protocol.CmdPing})
	_ = exec.Execute(protocol.Command{Name: protocol.CmdSet, Args: []string{"k1", "v1"}, Raw: "SET k1 v1"})
	_ = exec.Execute(protocol.Command{Name: protocol.CmdDel, Args: []string{"missing"}, Raw: "DEL missing"})
	_ = exec.Execute(protocol.Command{Name: protocol.CmdDel, Args: []string{"k1"}, Raw: "DEL k1"})

	if len(appender.commands) != 2 {
		t.Fatalf("append count = %d, want %d", len(appender.commands), 2)
	}
	if appender.commands[0].Name != protocol.CmdSet {
		t.Fatalf("first append command = %s, want %s", appender.commands[0].Name, protocol.CmdSet)
	}
	if appender.commands[1].Name != protocol.CmdDel {
		t.Fatalf("second append command = %s, want %s", appender.commands[1].Name, protocol.CmdDel)
	}
}

func TestExecutorAppendHookFailureReturnsError(t *testing.T) {
	appender := &mockAppender{err: errors.New("disk full")}
	exec := NewExecutor(storage.NewMemoryStorage(), appender)

	got := exec.Execute(protocol.Command{Name: protocol.CmdSet, Args: []string{"k1", "v1"}, Raw: "SET k1 v1"})
	want := "-ERR persistence append failed\r\n"
	if got != want {
		t.Fatalf("response = %q, want %q", got, want)
	}
}
