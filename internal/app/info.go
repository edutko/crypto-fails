package app

import "github.com/edutko/crypto-fails/pkg/app"

func GetInfo() app.Info {
	return app.Info{
		Version: Version,
		Config:  Config(),
		License: mgr.license,
	}
}

var Version = "0.0.0"
