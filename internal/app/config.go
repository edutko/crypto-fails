package app

import (
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/weakprng"
	"github.com/edutko/crypto-fails/pkg/app"
)

var cfg = app.Config{
	ExternalURL:    "http://localhost:8080/",
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

func Config() app.Config {
	return cfg
}

func LoadConfig() (app.Config, error) {
	lock.Lock()
	defer lock.Unlock()

	if cfgLoaded {
		return cfg, nil
	}

	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr != "" {
		cfg.ListenAddr = listenAddr
		cfg.ExternalURL = (&url.URL{Scheme: "http:", Host: listenAddr, Path: "/"}).String()
	}

	externalURL := os.Getenv("EXTERNAL_URL")
	if externalURL != "" {
		u, err := url.Parse(externalURL)
		if err != nil {
			return cfg, err
		}
		cfg.ExternalURL = u.String()
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

	cfgLoaded = true

	return cfg, nil
}

var (
	lock      sync.Mutex
	cfgLoaded bool
)
