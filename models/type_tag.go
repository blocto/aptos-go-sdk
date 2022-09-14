package models

import (
	"github.com/the729/lcs"
)

type TypeTag interface{}

type TypeTagBool struct{}
type TypeTagU8 struct{}
type TypeTagU64 struct{}
type TypeTagU128 struct{}
type TypeTagAddress struct{}
type TypeTagSinger struct{}
type TypeTagVector struct {
	TypeTag
}
type TypeTagStruct struct {
	Address    AccountAddress
	Module     string
	Name       string
	TypeParams []TypeTag
}

var _ = lcs.RegisterEnum(
	(*TypeTag)(nil),
	TypeTagBool{},
	TypeTagU8{},
	TypeTagU64{},
	TypeTagU128{},
	TypeTagAddress{},
	TypeTagSinger{},
	TypeTagVector{},
	TypeTagStruct{},
)
