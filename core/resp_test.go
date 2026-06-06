package core

import (
	"testing"
)

func TestDecodeSimpleString(t *testing.T) {
	cases := map[string]struct {
		input string
		expected string
		expected_delta int
		expected_error bool
	}{
		"simple string": {
			input: "+OK\r\n",
			expected: "OK",
			expected_delta: 5,
			expected_error: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value, delta, err := decodeSimpleString([]byte(tc.input))
			if err != nil {
				if !tc.expected_error {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if value != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, value)
			}
			if delta != tc.expected_delta {
				t.Errorf("expected %d, got %d", tc.expected_delta, delta)
			}
		})
	}
}

func TestDecodeError(t *testing.T) {
	cases := map[string]struct {
		input string
		expected string
		expected_delta int
		expected_error bool
	}{
		"simple error": {
			input: "-ERR unknown command\r\n",
			expected: "ERR unknown command",
			expected_delta: 22,
			expected_error: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value, delta, err := decodeError([]byte(tc.input))
			if err != nil {
				if !tc.expected_error {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if value != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, value)
			}
			if delta != tc.expected_delta {
				t.Errorf("expected %d, got %d", tc.expected_delta, delta)
			}
		})
	}
}

func TestDecodeInteger64(t *testing.T) {
	cases := map[string]struct {
		input string
		expected int64
		expected_delta int
		expected_error bool
	}{
		"simple integer": {
			input: ":1000\r\n",
			expected: int64(1000),
			expected_delta: 7,
			expected_error: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value, delta, err := decodeInteger64([]byte(tc.input))
			if err != nil {
				if !tc.expected_error {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if value != tc.expected {
				t.Errorf("expected %d, got %d", tc.expected, value)
			}
			if delta != tc.expected_delta {
				t.Errorf("expected %d, got %d", tc.expected_delta, delta)
			}
		})
	}
}

func TestDecodeBulkString(t *testing.T) {
	cases := map[string]struct {
		input string
		expected string
		expected_delta int
		expected_error bool
	}{
		"simple bulk string": {
			input: "$5\r\nhello\r\n",
			expected: "hello",
			expected_delta: 11,
			expected_error: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value, delta, err := decodeBulkString([]byte(tc.input))

			if err != nil {
				if !tc.expected_error {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if value != tc.expected {
				t.Errorf("expected %s, got %s", tc.expected, value)
			}
			if delta != tc.expected_delta {
				t.Errorf("expected %d, got %d", tc.expected_delta, delta)
			}
		})
	}
}

func TestDecodeArray(t *testing.T) {
	cases := map[string]struct {
		input string
		expected []interface{}
		expected_delta int
		expected_error bool
	}{
		"simple array": {
			input: "*2\r\n$5\r\nhello\r\n$5\r\nworld\r\n",
			expected: []interface{}{"hello", "world"},
			expected_delta: 26,
			expected_error: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			value, delta, err := decodeArray([]byte(tc.input))

			if err != nil {
				if !tc.expected_error {
					t.Errorf("unexpected error: %v", err)
				}
				return
			}
			if delta != tc.expected_delta {
				t.Errorf("expected %d, got %d", tc.expected_delta, delta)
			}
			
			for i := range tc.expected {
				if value[i] != tc.expected[i] {
					t.Errorf("expected %v, got %v", tc.expected, value)
				}
			}
		})
	}
}
