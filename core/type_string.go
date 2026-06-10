package core

import "strconv"

func deduceTypeEncoding(v string) (uint8, uint8) {
	var objectType uint8 = OBJ_TYPE_STRING
	if _, err := strconv.ParseInt(v, 10, 64); err == nil {
		return objectType, OBJ_ENCODING_INT
	}

	if len(v) <= 44 {
		return objectType, OBJ_ENCODING_EMBSTR
	}

	return objectType, OBJ_ENCODING_RAW
}
