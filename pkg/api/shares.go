package api

import (
	"github.com/edutko/crypto-fails/pkg/share"
)

type SharesResponse struct {
	Links []share.Link `json:"links"`
}
