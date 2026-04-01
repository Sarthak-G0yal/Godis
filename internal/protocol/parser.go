package protocol

import (
	"errors"
	"fmt"
	"strings"
)

var ErrEmptyCommand = errors.New("empty command")
var ErrUnknownCommand = errors.New("unknown command")
var ErrInvalidArity = errors.New("invalid number of arguments")

func ParseLine(raw string) (Command, error) {
	line := strings.TrimSpace(raw)
	if line == "" {
		return Command{}, ErrEmptyCommand
	}
	parts := strings.Fields(line)
	if len(parts) == 0 {
		return Command{}, ErrEmptyCommand
	}
	name := Name(strings.ToUpper(parts[0]))

	switch name {
	case CmdPing:
		if len(parts) != 1 {
			return Command{}, fmt.Errorf("%w: PING command takes no arguments", ErrInvalidArity)
		}
		return Command{Name: CmdPing, Raw: line}, nil
	case CmdGet:
		if len(parts) != 2 {
			return Command{}, fmt.Errorf("%w: GET command takes exactly one argument", ErrInvalidArity)
		}
		return Command{Name: CmdGet, Args: []string{parts[1]}, Raw: line}, nil
	case CmdSet:
		if len(parts) < 3 {
			return Command{}, fmt.Errorf("%w: SET command takes at least two arguments", ErrInvalidArity)
		}
		key := parts[1]
		value := strings.Join(parts[2:], " ")
		return Command{Name: CmdSet, Args: []string{key, value}, Raw: line}, nil
	case CmdDel:
		if len(parts) != 2 {
			return Command{}, fmt.Errorf("%w: DEL command takes exactly one argument", ErrInvalidArity)
		}
		return Command{Name: CmdDel, Args: []string{parts[1]}, Raw: line}, nil

	case CmdExists:
		if len(parts) != 2 {
			return Command{}, fmt.Errorf("%w: EXISTS command takes exactly one argument", ErrInvalidArity)
		}
		return Command{Name: CmdExists, Args: []string{parts[1]}, Raw: line}, nil

	default:
		return Command{}, fmt.Errorf("%w: %q", ErrUnknownCommand, name)
	}
}
