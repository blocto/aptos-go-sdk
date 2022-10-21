package models

import (
	"fmt"
	"strconv"
)

type Uint64 uint64

func (u *Uint64) UnmarshalJSON(b []byte) error {
	d := string(b)
	if len(d) < 3 {
		return fmt.Errorf("invalid Uint64")
	}
	d = d[1 : len(d)-1]
	v, err := strconv.ParseUint(d, 10, 64)
	if err != nil {
		return err
	}
	*u = Uint64(v)
	return nil
}
