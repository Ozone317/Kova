package core

import (
	"errors"
	"net"
)

func evalPING(args []string, conn net.Conn) error {
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

func EvalAndRespond(command *KovaCmd, conn net.Conn) error {
	switch command.Cmd {
	case "PING":
		return evalPING(command.Args, conn)
	default:
		return evalPING(command.Args, conn)
	}
}
