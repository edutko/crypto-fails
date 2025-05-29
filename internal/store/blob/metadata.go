package blob

import "time"

type Metadata struct {
	Key      string    `json:"key"`
	Modified time.Time `json:"modified"`
	Size     int64     `json:"size"`
}
