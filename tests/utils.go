package tests

import (
	"crypto/rand"
	"github.com/iden3/go-iden3-crypto/babyjub"
)

// RandomBytes creates slice of provided size filled with random values.
// It panics in case of error
func RandomBytes(size int) []byte {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		panic(err)
	}
	return buf
}

// RandomEdDSAPubKey returns random compressed BJJ public key
func RandomEdDSAPubKey() [32]byte {
	pk := babyjub.NewRandPrivKey()
	return pk.Public().Compress()
}
