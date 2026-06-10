package core

import (
	"errors"
)

// TYPE        ENCODINGS
// --------------------------------------------------------
// String      INT         when the value is a whole number
//             EMBSTR      when the string is ≤ 44 bytes
//             RAW         when the string is > 44 bytes

// List        LISTPACK    when the list is small (≤ 128 items, each ≤ 64 bytes)
//             QUICKLIST   when the list grows beyond those limits

// Hash        LISTPACK    when the hash is small (≤ 128 fields, each ≤ 64 bytes)
//             HASHTABLE   when the hash grows beyond those limits

// Set         LISTPACK    when the set is small and all values are non-integer
//             INTSET      when the set is small and all values are integers
//             HASHTABLE   when the set grows beyond the limits

// ZSet        LISTPACK    when the sorted set is small (≤ 128 items, each ≤ 64 bytes)
// (sorted)    SKIPLIST    when it grows beyond those limits

type Object struct {
	value        interface{}
	expiresAtMS  int64
	typeEncoding uint8 // type as the first 4 bits, encoding as the last 4 bits
}

const OBJ_TYPE_STRING = 0
const OBJ_TYPE_LIST = 1
const OBJ_TYPE_HASH = 2
const OBJ_TYPE_SET = 3
const OBJ_TYPE_ZSET = 4

const OBJ_ENCODING_INT = 0
const OBJ_ENCODING_EMBSTR = 1
const OBJ_ENCODING_RAW = 2
const OBJ_ENCODING_LISTPACK = 3
const OBJ_ENCODING_QUICKLIST = 4
const OBJ_ENCODING_HASHTABLE = 5
const OBJ_ENCODING_INTSET = 6
const OBJ_ENCODING_SKIPLIST = 7

func assertType(typeEncoding uint8, given_type uint8) error {
	if getType(typeEncoding) != given_type {
		return errors.New("the operation is not permitted on this type")
	}
	return nil
}

func assertEncoding(typeEncoding uint8, given_encoding uint8) error {
	if getEncoding(typeEncoding) != given_encoding {
		return errors.New("the opertaion is not permitted on this type")
	}
	return nil
}

func getType(typeEncoding uint8) uint8 {
	return (typeEncoding & 0b11110000) >> 4
}

func getEncoding(typeEncoding uint8) uint8 {
	return (typeEncoding & 0b00001111)
}
