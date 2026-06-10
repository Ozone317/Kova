package core

import (
	"bytes"
	"errors"
	"io"
	"strconv"
)

func evalPING(args []string) ([]byte, error) {
	if len(args) >= 2 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'ping' command")
	}

	var b []byte

	if len(args) == 0 {
		b = Encode("PONG", true)
	} else {
		b = Encode(args[0], false)
	}

	return b, nil
}

func evalSET(args []string) ([]byte, error) {
	if len(args) < 2 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'set' command")
	}

	var key string = args[0]
	var val string = args[1]
	var ttl int64 = -1

	objectType, objectEncoding := deduceTypeEncoding(val)

	for i := 2; i < len(args); i++ {
		if i < len(args) {
			if args[i] != "EX" && args[i] != "ex" {
				return nil, errors.New("(error) ERR syntax error")
			}
			i++
			if i >= len(args) {
				return nil, errors.New("(error) ERR syntax error")
			}
			time, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				return nil, errors.New("(error) ERR syntax error")
			}
			if time <= 0 {
				return nil, errors.New("(error) ERR invalid expire time in set")
			}
			ttl = time
		}
	}

	Put(key, NewObject(val, objectType, objectEncoding, ttl))

	return Encode("OK", true), nil
}

func evalGET(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'get' command")
	}

	val, ok := Get(args[0])
	if !ok {
		return Encode("(nil)", true), nil
	}

	return Encode(val.value, false), nil
}

func evalTTL(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'ttl' command")
	}

	return Encode(Ttl(args[0]), false), nil
}

func evalDEL(args []string) ([]byte, error) {
	if len(args) < 1 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'del' command")
	}

	return Encode(Del(args), false), nil
}

func evalEXPIRE(args []string) ([]byte, error) {
	if len(args) != 2 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'expire' command")
	}

	key := args[0]
	ttl, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return nil, errors.New("(error) ERR value is not an integer or out of range")
	}
	if ttl <= 0 {
		return nil, errors.New("(error) ERR invalid expire time in expire")
	}

	return Encode(Expire(key, ttl), false), nil
}

func evalBGREWRITEAOF(args []string) ([]byte, error) {
	if len(args) != 0 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'bgrewriteaof' command")
	}
	DumpAllAOF()
	return Encode("OK", true), nil
}

func evalINCR(args []string) ([]byte, error) {
	if len(args) != 1 {
		return nil, errors.New("(error) ERR wrong number of arguments for 'incr' command")
	}

	key := args[0]
	obj, _ := Get(key)

	if obj == nil {
		obj = NewObject("0", OBJ_TYPE_STRING, OBJ_ENCODING_INT, -1)
		Put(key, obj)
	}

	if err := assertType(obj.typeEncoding, OBJ_TYPE_STRING); err != nil {
		return nil, err
	}

	if err := assertEncoding(obj.typeEncoding, OBJ_ENCODING_INT); err != nil {
		return nil, err
	}

	val, err := strconv.ParseInt(obj.value.(string), 10, 64)
	if err != nil {
		return nil, errors.New("(error) ERR value is not an integer or out of range")
	}

	val++
	obj.value = strconv.FormatInt(val, 10)

	return Encode(val, false), nil
}

func EvalAndRespond(commands *KovaCmds, conn io.ReadWriter) error {
	var buf bytes.Buffer
	for _, command := range *commands {
		var (
			b   []byte
			err error
		)

		switch command.Cmd {
		case "PING":
			b, err = evalPING(command.Args)
		case "SET":
			b, err = evalSET(command.Args)
		case "GET":
			b, err = evalGET(command.Args)
		case "TTL":
			b, err = evalTTL(command.Args)
		case "DEL":
			b, err = evalDEL(command.Args)
		case "EXPIRE":
			b, err = evalEXPIRE(command.Args)
		case "BGREWRITEAOF":
			b, err = evalBGREWRITEAOF(command.Args)
		case "INCR":
			b, err = evalINCR(command.Args)
		default:
			err = errors.New("unknown command")
		}

		if err != nil {
			return err
		}

		buf.Write(b)
	}

	_, err := conn.Write(buf.Bytes())
	return err
}
