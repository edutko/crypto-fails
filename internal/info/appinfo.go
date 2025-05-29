package info

type Info struct {
	Version string `json:"version"`
}

func Initialize(version string) {
	info = Info{Version: version}
}

func GetAppInfo() Info {
	return info
}

var info = Info{
	Version: "0.0.0-dev",
}
