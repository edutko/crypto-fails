package app

type Info struct {
	Version string  `json:"version"`
	Config  Config  `json:"config"`
	License License `json:"license"`
}
