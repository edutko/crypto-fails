package app

func SetVersion(version string) {
	ver = version
}

func Version() string {
	return ver
}

var ver = "0.0.0"
