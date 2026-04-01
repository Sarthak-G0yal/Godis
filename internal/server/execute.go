package server

import (
	"errors"
	"fmt"
	"mini-redis/internal/protocol"
	"mini-redis/internal/storage"
)

type Executor struct {
	store storage.Storage
}

func NewExecutor(store storage.Storage) *Executor {
	return &Executor{store: store}
}

func (e *Executor) Execute(cmd protocol.Command) string {
	switch cmd.Name {
	case protocol.CmdPing:
		if err := expectArgs(cmd, 0); err != nil {
			return protocol.Error(err.Error())
		}
		return protocol.Pong()
	case protocol.CmdSet:
		if err := expectArgs(cmd, 2); err != nil {
			return protocol.Error(err.Error())
		}
		key := cmd.Args[0]
		value := cmd.Args[1]
		if err := e.store.Set(key, value); err != nil {
			return protocol.Error(mapStorageError(err))
		}
		return protocol.OK()
	case protocol.CmdGet:
		if err := expectArgs(cmd, 1); err != nil {
			return protocol.Error(err.Error())
		}
		key := cmd.Args[0]
		value, err := e.store.Get(key)
		if err != nil {
			if errors.Is(err, storage.ErrKeyNotFound) {
				return protocol.NilBulkString()
			}
			return protocol.Error(mapStorageError(err))
		}
		return protocol.BulkString(value)
	case protocol.CmdDel:
		if err := expectArgs(cmd, 1); err != nil {
			return protocol.Error(err.Error())
		}
		key := cmd.Args[0]
		if err := e.store.Delete(key); err != nil {
			if errors.Is(err, storage.ErrKeyNotFound) {
				return protocol.Integer(0)
			}
			return protocol.Error(mapStorageError(err))
		}
		return protocol.Integer(1)
	case protocol.CmdExists:
		if err := expectArgs(cmd, 1); err != nil {
			return protocol.Error(err.Error())
		}
		key := cmd.Args[0]
		return protocol.IntegerBool(e.store.Exists(key))
	default:
		return protocol.Error("unknown command")
	}
}

func expectArgs(cmd protocol.Command, want int) error {
	if len(cmd.Args) != want {
		return fmt.Errorf("invalid arity for %s", cmd.Name)
	}
	return nil
}

func mapStorageError(err error) string {
	switch {
	case errors.Is(err, storage.ErrKeyNotFound):
		return "key not found"
	case errors.Is(err, storage.ErrEmptyKey):
		return "empty key"
	default:
		return "internal storage error"
	}
}
