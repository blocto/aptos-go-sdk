package models

import (
	"github.com/the729/lcs"
)

type TransactionArgument interface{}

type TxArgU8 struct {
	U8 uint8
}

type TxArgU64 struct {
	U64 uint64
}

type TxArgU128 struct {
	Higher, Lower uint64
}

type TxArgBool struct {
	Bool bool
}

type TxArgAddress struct {
	Addr AccountAddress
}

type TxArgU8Vector struct {
	Bytes []byte
}

var _ = lcs.RegisterEnum(
	(*TransactionArgument)(nil),
	TxArgU8{},
	TxArgU64{},
	TxArgU128{},
	TxArgBool{},
	TxArgAddress{},
	TxArgU8Vector{},
)
