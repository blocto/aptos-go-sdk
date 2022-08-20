package models

import (
	"encoding/hex"
	"fmt"
	"strings"
)

type AccountAddress [32]byte

func (addr AccountAddress) PrefixZeroTrimmedHex() string {
	nonZeroIndex := 0
	for nonZeroIndex < 32 && addr[nonZeroIndex] == 0 {
		nonZeroIndex++
	}
	if nonZeroIndex == 32 {
		return "0x0"
	}

	hex := hex.EncodeToString(addr[nonZeroIndex:])
	return "0x" + strings.TrimPrefix(hex, "0")
}

func HexToAccountAddress(addr string) (AccountAddress, error) {
	addr = strings.TrimPrefix(addr, "0x")
	addrBytes, err := hex.DecodeString(addr)
	if err != nil {
		return [32]byte{}, err
	}

	length := len(addrBytes)
	if length > 32 {
		return [32]byte{}, fmt.Errorf("unexpected addr length: %d", length)
	}

	paddingBytes := make([]byte, 32-length)
	addrBytes = append(paddingBytes, addrBytes...)
	return *(*[32]byte)(addrBytes), nil
}
