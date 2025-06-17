package ecdsa

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"math/big"
)

func InsecureSignASN1(priv *ecdsa.PrivateKey, hash []byte) ([]byte, error) {
	c := priv.Curve
	N := c.Params().N
	if N.Sign() == 0 {
		return nil, errors.New("zero parameter")
	}

	// https://xkcd.com/221/
	b := bytes.Repeat([]byte{0x04}, (c.Params().N.BitLen()+7)/8)
	k := new(big.Int).SetBytes(b)

	kInv := new(big.Int).ModInverse(k, N)

	r, _ := c.ScalarBaseMult(k.Bytes())
	r.Mod(r, N)

	e := hashToInt(hash, c)
	s := new(big.Int).Mul(priv.D, r)
	s.Add(s, e)
	s.Mul(s, kInv)
	s.Mod(s, N)

	return asn1.Marshal(signature{r, s})
}

// hashToInt converts a hash value to an integer. Per FIPS 186-4, Section 6.4,
// we use the left-most bits of the hash to match the bit-length of the order of
// the curve. This also performs Step 5 of SEC 1, Version 2.0, Section 4.1.3.
func hashToInt(hash []byte, c elliptic.Curve) *big.Int {
	orderBits := c.Params().N.BitLen()
	orderBytes := (orderBits + 7) / 8
	if len(hash) > orderBytes {
		hash = hash[:orderBytes]
	}

	ret := new(big.Int).SetBytes(hash)
	excess := len(hash)*8 - orderBits
	if excess > 0 {
		ret.Rsh(ret, uint(excess))
	}
	return ret
}

type signature struct {
	R *big.Int
	S *big.Int
}
