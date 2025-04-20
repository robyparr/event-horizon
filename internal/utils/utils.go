package utils

import (
	"crypto/rand"
	"math/big"
)

const base62Chars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

func Token() string {
	b := make([]byte, 16)
	rand.Read(b)

	return encodeBase62(b)
}

func encodeBase62(b []byte) string {
	num := new(big.Int).SetBytes(b)

	var encoded []byte
	base := big.NewInt(62)
	zero := big.NewInt(0)
	mod := new(big.Int)

	for num.Cmp(zero) > 0 {
		num.DivMod(num, base, mod)
		encoded = append([]byte{base62Chars[mod.Int64()]}, encoded...)
	}

	return string(encoded)
}
