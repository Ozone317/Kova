package core

import (
	"errors"
	"fmt"
	"log"
)

/*
* This file will handle the RESP (REdis Serialization Protocol) encoding and decoding.
* Here are the rules for decoding:
* Prefix				Type				Example
* +						Simple String		+OK\r\n
* -						Error				-ERR Unknown command\r\n
* :						Integer				:1000\r\n
* $						Bulk String			$5\r\nhello\r\n
* *						Array				*2\r\n$3\r\nfoo\r\n$3\r\nbar\r\n

* Special cases:
* 1. Empty bulk string: $0\r\n\r\n
* 2. Null bulk string: $-1\r\n
* 3. Null array: *-1\r\n
 */

func Decode(data []byte) ([]interface{}, error) {
	if len(data) == 0 {
		return nil, errors.New("no data")
	}

	var i int = 0
	var values []interface{}

	for i < len(data) {
		value, delta, err := DecodeOne(data[i:])
		if err != nil {
			return nil, err
		}
		values = append(values, value)
		i += delta
	}

	return values, nil
}

func DecodeOne(data []byte) (interface{}, int, error) {
	if len(data) == 0 {
		return nil, 0, errors.New("no data")
	}

	switch data[0] {
	case '+':
		return decodeSimpleString(data)
	case '-':
		return decodeError(data)
	case ':':
		return decodeInteger64(data)
	case '$':
		return decodeBulkString(data)
	case '*':
		return decodeArray(data)
	default:
		return nil, 0, errors.New("unknown type")
	}
}

// reads a RESP-encoded simple string from data and
// returns the string, the delta, and the error
func decodeSimpleString(data []byte) (string, int, error) {
	// first character = '+
	pos := 1

	for ; data[pos] != '\r'; pos++ {

	}

	return string(data[1:pos]), pos + 2, nil // pos + 2 to skip the \r\n and start reading the next value
}

// reads a RESP-encoded error from data and
// returns the error string, the delta, and the error
func decodeError(data []byte) (string, int, error) {
	return decodeSimpleString(data)

}

// reads a RESP-encoded integer from data and
// returns the integer, the delta, and the error
func decodeInteger64(data []byte) (int64, int, error) {
	pos := 1
	var value int64 = 0

	for ; data[pos] != '\r'; pos++ {
		value = value*10 + int64(data[pos]-'0')
	}

	return value, pos + 2, nil
}

func decodeBulkString(data []byte) (string, int, error) {
	// first character = '$'
	pos := 1

	// second character = length of the string
	// example: $5\r\nhello\r\n
	// we need to read the characters until we encounter a \r

	len, delta := readLength(data[pos:])
	pos += delta

	// reading `len` bytes as string
	return string(data[pos : pos+len]), pos + len + 2, nil
}

func readLength(data []byte) (int, int) {
	var pos int = 0
	var length int = 0

	for pos := range data {
		b := data[pos]
		if !(b >= '0' && b <= '9') {
			return length, pos + 2
		}
		length = length*10 + int(b-'0')
	}

	return length, pos + 2
}

func decodeArray(data []byte) ([]interface{}, int, error) {
	// example of an RESP-encoded array: *2\r\n$5\r\nhello\r\n$3\r\nworld\r\n

	// first character = '*'
	pos := 1

	element_count, delta := readLength(data[pos:])
	pos += delta

	elements := make([]interface{}, element_count)

	for i := range elements {
		elem, delta, err := DecodeOne(data[pos:])
		if err != nil {
			return nil, 0, err
		}
		elements[i] = elem
		pos += delta
	}

	return elements, pos, nil
}

// Gets an array of bytes containing commands and converts it into an array of strings
func DecodeArrayString(data []byte) ([][]string, error) {
	values, err := Decode(data)
	if err != nil {
		return nil, err
	}
	all_tokens := make([][]string, 0)
	for _, value := range values {
		ts := value.([]interface{})
		tokens := make([]string, len(ts))
		for i := range ts {
			tokens[i] = ts[i].(string)
		}
		all_tokens = append(all_tokens, tokens)
		log.Println("tokens: ", tokens)
	}
	return all_tokens, nil
}

func Encode(value interface{}, isSimple bool) []byte {
	switch v := value.(type) {
	case string:
		if isSimple {
			return []byte(fmt.Sprintf("+%s\r\n", v))
		}
		return []byte(fmt.Sprintf("$%d\r\n%s\r\n", len(v), v))
	case int, int64:
		return []byte(fmt.Sprintf(":%d\r\n", v))
	}

	return []byte{}
}
