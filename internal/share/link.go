package share

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

type Link struct {
	Id         string    `json:"id,omitempty"`
	Key        string    `json:"key"`
	Expiration time.Time `json:"expiration,omitempty"`
	Signature  string    `json:"-"`
}

var (
	ErrExpired          = errors.New("expired")
	ErrInvalidSignature = errors.New("invalid signature")
	ErrNoSignature      = errors.New("no signature")

	DoesNotExpire  = time.Time{}
	AlreadyExpired = time.Unix(0, 0).UTC()
)

func NewLink(key string, expiration time.Time) Link {
	return Link{
		Key:        key,
		Expiration: expiration.UTC(),
	}
}

func ParseLink(v url.Values) Link {
	var exp time.Time
	if v.Get("exp") == "" {
		exp = DoesNotExpire
	} else if expInt, err := strconv.ParseInt(v.Get("exp"), 10, 64); err == nil {
		exp = time.Unix(expInt, 0).UTC()
	} else {
		exp = AlreadyExpired
	}
	return Link{
		Key:        v.Get("key"),
		Expiration: exp,
		Signature:  v.Get("sig"),
	}
}

func (l Link) QueryString() string {
	return urlValues(l).Encode()
}

func (l Link) SignedQueryString(secret []byte) string {
	sig := l.signWithSecret(secret)
	l.Signature = base64.RawURLEncoding.EncodeToString(sig)
	return urlValues(l).Encode()
}

func (l Link) Verify(secret []byte) error {
	if l.Signature == "" {
		return ErrNoSignature
	}

	sig, err := base64.RawURLEncoding.DecodeString(l.Signature)
	if err != nil {
		return ErrInvalidSignature
	}

	computedSig := l.signWithSecret(secret)
	if !bytes.Equal(sig, computedSig[:]) {
		return ErrInvalidSignature
	}

	if !l.Expiration.IsZero() && l.Expiration.Before(timeNow()) {
		return ErrExpired
	}

	return nil
}

func (l Link) signWithSecret(secret []byte) []byte {
	signedValues := urlValues(l)
	signedValues.Del("sig")
	decoded, _ := url.QueryUnescape(signedValues.Encode())
	sig := sha256.Sum256(append(secret, []byte(decoded)...))
	return sig[:]
}

func urlValues(l Link) url.Values {
	v := make(url.Values)
	v.Set("key", l.Key)
	if !l.Expiration.IsZero() {
		v.Set("exp", fmt.Sprintf("%d", l.Expiration.Unix()))
	}
	if l.Signature != "" {
		v.Set("sig", l.Signature)
	}
	return v
}

var timeNow = time.Now
