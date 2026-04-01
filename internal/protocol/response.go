package protocol

import (
	"fmt"
	"strconv"
)

const crlf = "\r\n"

func SimpleString(msg string) string {
	return "+" + msg + crlf
}

func Error(msg string) string {
	return "-ERR " + msg + crlf
}

func Integer(num int64) string {
	return ":" + strconv.FormatInt(num, 10) + crlf
}

func IntegerBool(ok bool) string {
	if ok {
		return Integer(1)
	}
	return Integer(0)
}

func BulkString(value string) string {
	return fmt.Sprintf("$%d%s%s%s", len(value), crlf, value, crlf)
}

func NilBulkString() string {
	return "$-1" + crlf
}

func Pong() string {
	return SimpleString("PONG")
}

func OK() string {
	return SimpleString("OK")
}
