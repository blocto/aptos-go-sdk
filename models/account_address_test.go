package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodHex(t *testing.T) {
	t.Run("AptosReservedAddresses", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			aptosReservedAddr := fmt.Sprintf("0x%v", i)
			addr, err := HexToAccountAddress(aptosReservedAddr)
			assert.NoError(t, err)
			assert.Equal(t, aptosReservedAddr, addr.ToHex())
		}

		aptosReservedAddr := "0xa"
		addr, err := HexToAccountAddress(aptosReservedAddr)
		assert.NoError(t, err)
		assert.Equal(t, aptosReservedAddr, addr.ToHex())
	})

	t.Run("address", func(t *testing.T) {
		addr := "0x1111111111111111111111111111111111111111111111111111111111111111"
		accountAddr, err := HexToAccountAddress(addr)
		assert.NoError(t, err)

		assert.Equal(t, addr, accountAddr.ToHex())
	})

	t.Run("addressWithZeroAfter0xPrefix", func(t *testing.T) {
		addr := "0x0111111111111111111111111111111111111111111111111111111111111111"
		accountAddr, err := HexToAccountAddress(addr)
		assert.NoError(t, err)

		assert.Equal(t, addr, accountAddr.ToHex())
	})
}
