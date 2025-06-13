package app

import (
	"time"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/weakprng"
)

type Config struct {
	ExternalURL    string `json:"externalURL"`
	ListenAddr     string `json:"listenAddr"`
	StorageRootDir string `json:"storageRootDir"`
	WebRootDir     string `json:"webRootDir"`

	FileSizeLimit       int64              `json:"fileSizeLimit"`
	FileEncryptionMode  crypto.Mode        `json:"fileEncryptionMode"`
	LeakEncryptedFiles  bool               `json:"leakEncryptedFiles"`
	TweakEncryptedFiles bool               `json:"tweakEncryptedFiles"`
	WeakPRNGAlgorithm   weakprng.Algorithm `json:"weakPRNGAlgorithm"`

	SessionDuration   time.Duration `json:"sessionDuration"`
	ShareLinkDuration time.Duration `json:"shareLinkDuration"`
}
