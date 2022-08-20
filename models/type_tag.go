package models

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/the729/lcs"
)

type TypeTag interface {
	ToString() string
}

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

func (t TypeTagBool) ToString() string {
	return "Bool"
}

func (t TypeTagU8) ToString() string {
	return "U8"
}

func (t TypeTagU64) ToString() string {
	return "U64"
}

func (t TypeTagU128) ToString() string {
	return "U128"
}

func (t TypeTagAddress) ToString() string {
	return "Address"
}

func (t TypeTagSinger) ToString() string {
	return "Singer"
}

func (t TypeTagVector) ToString() string {
	return fmt.Sprintf("vector<%s>", t.TypeTag.ToString())
}

func (t TypeTagStruct) ToString() string {
	structType := fmt.Sprintf("%s::%s::%s", "0x"+hex.EncodeToString(t.Address[:]), t.Module, t.Name)
	if len(t.TypeParams) == 0 {
		return structType
	}

	var types []string
	for _, p := range t.TypeParams {
		types = append(types, p.ToString())
	}

	return fmt.Sprintf("%s<%s>", structType, strings.Join(types, ","))
}
