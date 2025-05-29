package job

import "time"

type Descriptor struct {
	Id         string    `json:"id"`
	StartedAt  time.Time `json:"startedAt"`
	FinishedAt time.Time `json:"finishedAt"`
	Cancelled  bool      `json:"cancelled"`
	Errors     []error   `json:"errors"`
}

type Status string

const (
	StatusNotStarted Status = "not started"
	StatusStarted    Status = "started"
	StatusCanceled   Status = "canceled"
	StatusFailed     Status = "failed"
	StatusCompleted  Status = "completed"
)

func (d Descriptor) Status() Status {
	if d.StartedAt.IsZero() {
		return StatusNotStarted
	}
	if d.FinishedAt.IsZero() {
		return StatusStarted
	}
	if d.Cancelled {
		return StatusCanceled
	}
	if len(d.Errors) > 0 {
		return StatusFailed
	}
	return StatusCompleted
}
