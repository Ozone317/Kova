package core

import (
	"errors"
	"io"
	"strconv"
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

func evalSET(args []string, conn io.ReadWriter) error {
	if len(args) < 2 {
		return errors.New("ERR wrong number of arguments for 'set' command")
	}

	var key string = args[0]
	var val interface{} = args[1]
	var ttl int64 = -1

	for i := 2; i < len(args); i++ {
		if i < len(args) {
			if args[i] != "EX" && args[i] != "ex" {
				return errors.New("ERR syntax error")
			}
			i++
			if i >= len(args) {
				return errors.New("ERR syntax error")
			}
			time, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return errors.New("ERR syntax error")
			}
			if time <= 0 {
				return errors.New("ERR invalid expire time in set")
			}
			ttl = time
		}
	}

	Put(key, val, ttl)
	_, err := conn.Write(Encode("OK", true))
	if err != nil {
		return err
	}
	return nil
}

func evalGET(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}

	val, ok := Get(args[0])
	if !ok {
		_, err := conn.Write(Encode("(nil)", true))
		return err
	}

	_, err := conn.Write(Encode(val.value, false))
	return err
}

func evalTTL(args []string, conn io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}

	_, err := conn.Write(Encode(Ttl(args[0]), false))
	return err
}

func EvalAndRespond(command *KovaCmd, conn io.ReadWriter) error {
	switch command.Cmd {
	case "PING":
		return evalPING(command.Args, conn)
	case "SET":
		return evalSET(command.Args, conn)
	case "GET":
		return evalGET(command.Args, conn)
	case "TTL":
		return evalTTL(command.Args, conn)
	default:
		return evalPING(command.Args, conn)
	}
}
