package core

import (
	"errors"
	"io"
)

func evalPING(args []string, conn io.ReadWriter) error {
	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	var b []byte

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	_, err := conn.Write(b)
	return err
}

func EvalAndRespond(command *KovaCmd, conn io.ReadWriter) error {
	switch command.Cmd {
	case "PING":
		return evalPING(command.Args, conn)
	default:
		return evalPING(command.Args, conn)
	}
}
