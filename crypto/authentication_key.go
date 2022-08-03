package crypto

import (
	"golang.org/x/crypto/sha3"
)

func SingleSignerAuthKey(publicKey [32]byte) [32]byte {
	return sha3.Sum256(append(publicKey[:], 0x00))
}

func MultiSignerAuthKey(publicKeys ...[32]byte) [32]byte {
	length := len(publicKeys)*32 + 2
	rawKey := make([]byte, length)
	index := 0
	for _, publicKey := range publicKeys {
		copy(rawKey[index:index+32], publicKey[:])
		index += 32
	}

	rawKey[len(rawKey)-2] = byte(len(publicKeys))
	rawKey[len(rawKey)-1] = 0x01
	return sha3.Sum256(rawKey)
}
