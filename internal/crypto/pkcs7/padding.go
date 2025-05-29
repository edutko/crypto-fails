package pkcs7

import (
	"bytes"
	"errors"
)

var ErrBadPadding = errors.New("bad padding")

func Pad(data []byte, blockSize int) []byte {
	padLen := blockSize - len(data)%blockSize
	padding := bytes.Repeat([]byte{byte(padLen)}, padLen)
	return append(data, padding...)
}

func Unpad(data []byte, blockSize int) ([]byte, error) {
	padByte := data[len(data)-1]
	padLen := int(padByte)

	if padLen < 1 || padLen > blockSize {
		return nil, ErrBadPadding
	}
	for i := len(data) - 1; i >= len(data)-padLen; i-- {
		if data[i] != padByte {
			return nil, ErrBadPadding
		}
	}

	end := len(data) - padLen
	return data[0:end], nil
}
