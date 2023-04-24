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

func (addr AccountAddress) ToHex() string {
	/* According to Aptos docs:
	 * https://github.com/aptos-labs/aptos-core/blob/main/aptos-move/framework/aptos-framework/doc/account.md
	 * The system reserved addresses is 0x1 / 0x2 / 0x3 / 0x4 / 0x5 / 0x6 / 0x7 / 0x8 / 0x9 / 0xa
	 * These addreeses would be return as string 0x1 / 0x2 / 0x3 / 0x4 / 0x5 / 0x6 / 0x7 / 0x8 / 0x9 / 0xa
	 * Other addresses would be return as string 0x + hex string
	 */
	nonZeroIndex := 0
	for nonZeroIndex < 31 && addr[nonZeroIndex] == 0 {
		nonZeroIndex++
	}
	if nonZeroIndex == 31 {
		return fmt.Sprintf("0x%x", addr[31])
	}

	return "0x" + hex.EncodeToString(addr[:])
}

func HexToAccountAddress(addr string) (AccountAddress, error) {
	addr = strings.TrimPrefix(addr, "0x")
	if len(addr)%2 == 1 {
		addr = "0" + addr
	}
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
