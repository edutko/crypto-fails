package config

import (
	"crypto/rand"
	"os"
	"time"

	"github.com/MicahParks/jwkset"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/jwks"
)

type Config struct {
	ListenAddr     string `json:"listenAddr"`
	StorageRootDir string `json:"storageRootDir"`
	WebRootDir     string `json:"webRootDir"`

	SessionDuration     time.Duration             `json:"sessionDuration"`
	CookieEncryptionKey []byte                    `json:"cookieEncryptionKey"`
	JWTSigningKey       []byte                    `json:"jwtSigningKey"`
	JWKS                jwkset.Storage            `json:"-"`
	JWTSigningKeys      map[string]crypto.KeyPair `json:"-"`

	FileSizeLimit          int64         `json:"fileSizeLimit"`
	ShareLinkDuration      time.Duration `json:"shareLinkDuration"`
	ShareLinkSigningSecret []byte        `json:"shareLinkSigningSecret"`

	LeakEncryptedFiles bool `json:"leakEncryptedFiles"`
}

func Load() Config {
	return cfg
}

func SessionDuration() time.Duration {
	return cfg.SessionDuration
}

func CookieEncryptionKey() []byte {
	return cfg.CookieEncryptionKey
}

func JWTSigningKey() []byte {
	return cfg.JWTSigningKey
}

func JWKS() jwkset.Storage {
	return cfg.JWKS
}

func JWTSigningKeys() map[string]crypto.KeyPair {
	return cfg.JWTSigningKeys
}

func MaxFileSize() int64 {
	return cfg.FileSizeLimit
}

func ShareLinkDuration() time.Duration {
	return cfg.ShareLinkDuration
}

func ShareLinkSigningSecret() []byte {
	return cfg.ShareLinkSigningSecret
}

func defaultConfig() Config {
	listenAddr := os.Getenv("LISTEN_ADDR")
	if listenAddr == "" {
		listenAddr = "localhost:8080"
	}

	conf := Config{
		ListenAddr:     listenAddr,
		StorageRootDir: "data",
		WebRootDir:     "web/static",

		SessionDuration:   6 * time.Hour,
		ShareLinkDuration: 15 * 24 * time.Hour,

		FileSizeLimit:      100 * 1024 * 1024,
		LeakEncryptedFiles: true,
	}

	conf.CookieEncryptionKey = make([]byte, 32)
	_, err := rand.Read(conf.CookieEncryptionKey)
	if err != nil {
		panic(err)
	}

	conf.JWKS, conf.JWTSigningKeys = jwks.InitializeKeys()
	conf.JWTSigningKey = []byte("JWTSecretKeyDontUseInProduction!")

	conf.ShareLinkSigningSecret = make([]byte, 16)
	_, err = rand.Read(conf.ShareLinkSigningSecret)
	if err != nil {
		panic(err)
	}

	return conf
}

func InitializeForTesting(c Config) {
	cfg = c
}

var cfg = defaultConfig()
