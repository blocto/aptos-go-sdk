package models

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
)

// Uint64 represents a uint64 value for JSON string format.
type Uint64 uint64

func (u *Uint64) UnmarshalJSON(b []byte) error {
	b = bytes.Trim(b, "\"")
	s := strings.TrimSpace(string(b))
	if s == "" {
		*u = 0
		return nil
	}
	v, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return err
	}
	*u = Uint64(v)
	return nil
}

func (u Uint64) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"%d\"", u)), nil
}
