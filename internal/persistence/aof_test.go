package persistence

import (
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"mini-redis/internal/protocol"
	"mini-redis/internal/storage"
)

func TestAOFAppendWritesOnlyMutatingCommands(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	aof, err := NewAOF(path, 0)
	if err != nil {
		t.Fatalf("NewAOF failed: %v", err)
	}
	t.Cleanup(func() { _ = aof.Close() })

	if err := aof.Append(protocol.Command{Name: protocol.CmdPing, Raw: "PING"}); err != nil {
		t.Fatalf("append PING failed: %v", err)
	}
	if err := aof.Append(protocol.Command{Name: protocol.CmdSet, Raw: "SET name mini-redis"}); err != nil {
		t.Fatalf("append SET failed: %v", err)
	}
	if err := aof.Append(protocol.Command{Name: protocol.CmdDel, Raw: "DEL name"}); err != nil {
		t.Fatalf("append DEL failed: %v", err)
	}

	if err := aof.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	gotBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	got := string(gotBytes)
	want := "SET name mini-redis\nDEL name\n"
	if got != want {
		t.Fatalf("aof contents = %q, want %q", got, want)
	}

	if strings.Contains(got, "PING") {
		t.Fatalf("AOF should not contain non-write command, got %q", got)
	}
}

func TestAOFAppendFormatsWhenRawEmpty(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	aof, err := NewAOF(path, 0)
	if err != nil {
		t.Fatalf("NewAOF failed: %v", err)
	}
	t.Cleanup(func() { _ = aof.Close() })

	err = aof.Append(protocol.Command{
		Name: protocol.CmdSet,
		Args: []string{"msg", "hello world"},
	})
	if err != nil {
		t.Fatalf("append failed: %v", err)
	}

	if err := aof.Close(); err != nil {
		t.Fatalf("close failed: %v", err)
	}

	gotBytes, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	got := string(gotBytes)
	want := "SET msg hello world\n"
	if got != want {
		t.Fatalf("aof contents = %q, want %q", got, want)
	}
}

func TestAOFReplayAppliesWritesAndSkipsMalformed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	seed := strings.Join([]string{
		"SET name mini redis",
		"PING",
		"SET feature on",
		"DEL missing",
		"BROKEN ???",
		"DEL feature",
		"",
	}, "\n")

	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	aof, err := NewAOF(path, 0)
	if err != nil {
		t.Fatalf("NewAOF failed: %v", err)
	}
	t.Cleanup(func() { _ = aof.Close() })

	store := storage.NewMemoryStorage()
	if err := aof.Replay(store); err != nil {
		t.Fatalf("Replay failed: %v", err)
	}

	got, err := store.Get("name")
	if err != nil {
		t.Fatalf("Get(name) failed: %v", err)
	}
	if got != "mini redis" {
		t.Fatalf("name = %q, want %q", got, "mini redis")
	}

	if store.Exists("feature") {
		t.Fatalf("feature should have been deleted by replay")
	}
}

func TestAOFPeriodicSyncLoopRuns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	aof, err := NewAOF(path, 10*time.Millisecond)
	if err != nil {
		t.Fatalf("NewAOF failed: %v", err)
	}
	t.Cleanup(func() { _ = aof.Close() })

	var syncCount atomic.Int64

	aof.mu.Lock()
	aof.syncHook = func() { syncCount.Add(1) }
	aof.mu.Unlock()

	if err := aof.Append(protocol.Command{Name: protocol.CmdSet, Raw: "SET k v"}); err != nil {
		t.Fatalf("append failed: %v", err)
	}

	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if syncCount.Load() > 0 {
			break
		}
		time.Sleep(5 * time.Millisecond)
	}

	if syncCount.Load() == 0 {
		t.Fatalf("expected periodic sync to run at least once")
	}
}

func TestAOFCloseIsIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "appendonly.aof")

	aof, err := NewAOF(path, 0)
	if err != nil {
		t.Fatalf("NewAOF failed: %v", err)
	}

	if err := aof.Close(); err != nil {
		t.Fatalf("first close failed: %v", err)
	}

	if err := aof.Close(); err != nil {
		t.Fatalf("second close failed: %v", err)
	}
}
