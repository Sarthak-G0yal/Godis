package server

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"mini-redis/internal/protocol"
	"mini-redis/internal/storage"
	"net"
	"sync"
)

type Server struct {
	addr string
	exec *Executor

	mu       sync.Mutex
	listener net.Listener
	closing  bool

	connMu sync.Mutex
	conns  map[net.Conn]struct{}

	wg sync.WaitGroup
}

func New(addr string, store storage.Storage) *Server {
	return NewWithAppender(addr, store, nil)
}

func NewWithAppender(addr string, store storage.Storage, appender CommandAppender) *Server {
	return &Server{
		addr:  addr,
		exec:  NewExecutor(store, appender),
		conns: make(map[net.Conn]struct{}),
	}
}

func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.addr)
	if err != nil {
		return err
	}
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		_ = ln.Close()
		return context.Canceled
	}
	s.listener = ln
	s.mu.Unlock()

	log.Printf("mini-redis listening on %s", s.addr)

	for {
		conn, err := ln.Accept()
		if err != nil {
			if s.isClosing() {
				return nil
			}
			var ne net.Error
			if errors.As(err, &ne) && ne.Temporary() {
				log.Printf("temporary accept error: %v", err)
				continue
			}
			return err
		}
		s.trackConn(conn)
		s.wg.Add(1)
		go s.handleConnections(conn)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return nil
	}
	s.closing = true
	ln := s.listener
	s.mu.Unlock()
	if ln != nil {
		_ = ln.Close()
	}
	s.closeAllConns()
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (s *Server) handleConnections(conn net.Conn) {
	defer s.wg.Done()
	defer s.untrackConn(conn)
	defer conn.Close()
	reader := bufio.NewReader(conn)

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) || s.isClosing() {
				return
			}
			log.Printf("read error from %s: %v", conn.RemoteAddr(), err)
			_, _ = conn.Write([]byte(protocol.Error("read error")))
			return
		}
		cmd, err := protocol.ParseLine(line)
		if err != nil {
			_, _ = conn.Write([]byte(protocol.Error(err.Error())))
			continue
		}
		res := s.exec.Execute(cmd)
		if _, err := conn.Write([]byte(res)); err != nil {
			if !s.isClosing() {
				log.Printf("write error to %s: %v", conn.RemoteAddr(), err)
			}
			return
		}
	}
}

func (s *Server) isClosing() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.closing
}

func (s *Server) trackConn(conn net.Conn) {
	s.connMu.Lock()
	s.conns[conn] = struct{}{}
	s.connMu.Unlock()
}

func (s *Server) untrackConn(conn net.Conn) {
	s.connMu.Lock()
	delete(s.conns, conn)
	s.connMu.Unlock()
}

func (s *Server) closeAllConns() {
	s.connMu.Lock()
	defer s.connMu.Unlock()
	for conn := range s.conns {
		_ = conn.Close()
	}
}
