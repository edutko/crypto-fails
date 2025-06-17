package config

import (
	"net/url"
	"time"

	"github.com/edutko/crypto-fails/internal/app"
	"github.com/edutko/crypto-fails/internal/crypto"
)

func BaseURL() *url.URL {
	u, err := url.Parse(app.Config().ExternalURL)
	if err != nil {
		panic(err)
	}
	return u
}

func FileEncryptionMode() crypto.Mode {
	return app.Config().FileEncryptionMode
}

func MaxFileSize() int64 {
	return app.Config().FileSizeLimit
}

func SessionDuration() time.Duration {
	return app.Config().SessionDuration
}

func ShareLinkDuration() time.Duration {
	return app.Config().ShareLinkDuration
}
