package api

import "github.com/edutko/crypto-fails/pkg/blob"

type FilesMetadataResponse struct {
	Files []blob.Metadata `json:"files"`
}
