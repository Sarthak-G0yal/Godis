package persistence

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"mini-redis/internal/protocol"
	"mini-redis/internal/storage"
	"os"
	"strings"
	"sync"
	"time"
)

var ErrAOFClosed = errors.New("aof is closed")

type AOF struct {
	file *os.File

	mu     sync.Mutex
	closed bool

	syncInterval time.Duration
	stopCh       chan struct{}
	doneCh       chan struct{}
	closeOnce    sync.Once

	syncHook func()
}

func NewAOF(path string, syncInterval time.Duration) (*AOF, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open aof: %w", err)
	}

	a := &AOF{
		file:         f,
		syncInterval: syncInterval,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}
	if syncInterval > 0 {
		go a.syncLoop()
	} else {
		close(a.doneCh)
	}
	return a, nil
}

func (a *AOF) Append(cmd protocol.Command) error {
	if !cmd.IsWrite() {
		return nil
	}
	line, err := formatCommand(cmd)
	if err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return ErrAOFClosed
	}
	if _, err := a.file.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("append aof: %w", err)
	}
	return nil
}

func (a *AOF) Replay(store storage.Storage) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return ErrAOFClosed
	}
	if _, err := a.file.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("seek aof start: %w", err)
	}
	scanner := bufio.NewScanner(a.file)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	lineNo := 0
	for scanner.Scan() {
		lineNo++
		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {
			continue
		}
		cmd, err := protocol.ParseLine(raw)
		if err != nil {
			log.Printf("aof replay: skip malformed line %d: %v", lineNo, err)
			continue
		}
		if !cmd.IsWrite() {
			continue
		}
		switch cmd.Name {
		case protocol.CmdSet:
			if len(cmd.Args) != 2 {
				log.Printf("aof replay: skip malformed set command at line %d: wrong number of args", lineNo)
			}
			if err := store.Set(cmd.Args[0], cmd.Args[1]); err != nil {
				log.Printf("aof replay: failed to execute set command at line %d: %v", lineNo, err)
			}
		case protocol.CmdDel:
			if len(cmd.Args) != 1 {
				log.Printf("aof replay: skip invalid DEL on line %d", lineNo)
				continue
			}
			if err := store.Delete(cmd.Args[0]); err != nil && !errors.Is(err, storage.ErrKeyNotFound) {
				log.Printf("aof replay: DEL failed on line %d: %v", lineNo, err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("scan aof: %w", err)
	}

	if _, err := a.file.Seek(0, io.SeekEnd); err != nil {
		return fmt.Errorf("seek aof end: %w", err)
	}

	return nil
}

func (a *AOF) Close() error {
    var closeErr error

    a.closeOnce.Do(func() {
        if a.syncInterval > 0 {
            close(a.stopCh)
            <-a.doneCh
        }

        a.mu.Lock()
        defer a.mu.Unlock()

        if a.closed {
            return
        }

        if err := a.syncUnderLock(); err != nil {
            closeErr = err
        }

        if err := a.file.Close(); err != nil {
            if closeErr == nil {
                closeErr = err
            } else {
                closeErr = errors.Join(closeErr, err)
            }
        }

        a.closed = true
    })

    return closeErr
}

func (a *AOF) syncLoop() {
    ticker := time.NewTicker(a.syncInterval)
    defer ticker.Stop()
    defer close(a.doneCh)

    for {
        select {
        case <-ticker.C:
            if err := a.sync(); err != nil {
                log.Printf("aof periodic sync failed: %v", err)
            }
        case <-a.stopCh:
            return
        }
    }
}

func (a *AOF) sync() error {
    a.mu.Lock()
    defer a.mu.Unlock()
    return a.syncUnderLock()
}

func (a *AOF) syncUnderLock() error {
    if a.closed {
        return nil
    }

    if err := a.file.Sync(); err != nil {
        return fmt.Errorf("sync aof: %w", err)
    }

    if a.syncHook != nil {
        a.syncHook()
    }

    return nil
}

func formatCommand(cmd protocol.Command) (string, error) {
    if raw := strings.TrimSpace(cmd.Raw); raw != "" {
        return raw, nil
    }

    switch cmd.Name {
    case protocol.CmdSet:
        if len(cmd.Args) != 2 {
            return "", fmt.Errorf("invalid SET command args: %d", len(cmd.Args))
        }
        return fmt.Sprintf("%s %s %s", protocol.CmdSet, cmd.Args[0], cmd.Args[1]), nil

    case protocol.CmdDel:
        if len(cmd.Args) != 1 {
            return "", fmt.Errorf("invalid DEL command args: %d", len(cmd.Args))
        }
        return fmt.Sprintf("%s %s", protocol.CmdDel, cmd.Args[0]), nil

    default:
        return "", fmt.Errorf("unsupported command for append: %s", cmd.Name)
    }
}