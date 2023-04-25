package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTodHex(t *testing.T) {
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
