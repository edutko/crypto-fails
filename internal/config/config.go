package config

import (
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/weakprng"
)

var cfg = Config{
	ListenAddr:     "localhost:8080",
	StorageRootDir: "data",
	WebRootDir:     "web/static",

	FileSizeLimit:       100 * 1024 * 1024,
	FileEncryptionMode:  crypto.ModeCTR,
	LeakEncryptedFiles:  true,
	TweakEncryptedFiles: true,
	WeakPRNGAlgorithm:   weakprng.XORShift128p,

	SessionDuration:   6 * time.Hour,
	ShareLinkDuration: 15 * 24 * time.Hour,
}

func Load() Config {
	lock.Lock()
	defer lock.Unlock()

	if cfg.loaded {
		return cfg
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr != "" {
		cfg.ListenAddr = listenAddr
	}

	switch strings.ToLower(os.Getenv("LEAK_ENCRYPTED_FILES")) {
	case "true", "yes", "1":
		cfg.LeakEncryptedFiles = true
	case "false", "no", "0":
		cfg.LeakEncryptedFiles = false
	}

	switch strings.ToLower(os.Getenv("TWEAK_ENCRYPTED_FILES")) {
	case "true", "yes", "1":
		cfg.TweakEncryptedFiles = true
	case "false", "no", "0":
		cfg.TweakEncryptedFiles = false
	}

	mode := os.Getenv("FILE_ENCRYPTION_MODE")
	switch strings.ToLower(mode) {
	case "ctr":
		cfg.FileEncryptionMode = crypto.ModeCTR
	case "gcm":
		cfg.FileEncryptionMode = crypto.ModeGCM
	}

	prng := os.Getenv("WEAK_PRNG")
	if prng != "" {
		cfg.WeakPRNGAlgorithm = weakprng.Algorithm(prng)
	}

	cfg.loaded = true

	return cfg
}

func MaxFileSize() int64 {
	return cfg.FileSizeLimit
}

func SessionDuration() time.Duration {
	return cfg.SessionDuration
}

func ShareLinkDuration() time.Duration {
	return cfg.ShareLinkDuration
}

func InitializeForTesting(c Config) {
	cfg = c
}

type Config struct {
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

	loaded bool
}

var lock sync.Mutex
