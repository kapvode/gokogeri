package gokogeri

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/kapvode/gokogeri/internal/sidekiq"
)

// redisJob represents the job data as it is encoded in Redis.
type redisJob struct {
	Class string        `json:"class"`
	Queue string        `json:"queue"`
	Args  []interface{} `json:"args,omitempty"`
	Retry bool          `json:"retry"`

	JobID      string  `json:"jid"`
	CreatedAt  float64 `json:"created_at"`
	EnqueuedAt float64 `json:"enqueued_at"`
}

type Job struct {
	enc redisJob

	createdAt  time.Time
	enqueuedAt time.Time
}

func newJobFromJSON(data []byte) (*Job, error) {
	job := &Job{}
	err := json.Unmarshal(data, &job.enc)
	if err != nil {
		return nil, fmt.Errorf("decoding job json: %v", err)
	}
	job.createdAt = sidekiq.ToTime(job.enc.CreatedAt)
	job.enqueuedAt = sidekiq.ToTime(job.enc.EnqueuedAt)
	return job, nil
}

func (j *Job) ID() string {
	return j.enc.JobID
}

func (j *Job) SetID(id string) *Job {
	j.enc.JobID = id
	return j
}

// Class returns the Ruby class that implements the job.
func (j *Job) Class() string {
	return j.enc.Class
}

func (j *Job) SetClass(c string) *Job {
	j.enc.Class = c
	return j
}

func (j *Job) Queue() string {
	return j.enc.Queue
}

func (j *Job) SetQueue(q string) *Job {
	j.enc.Queue = q
	return j
}

func (j *Job) Args() []interface{} {
	return j.enc.Args
}

func (j *Job) SetArgs(args []interface{}) *Job {
	j.enc.Args = args
	return j
}

func (j *Job) CreatedAt() time.Time {
	return j.createdAt
}

func (j *Job) SetCreatedAt(t time.Time) *Job {
	j.createdAt = t
	return j
}

func (j *Job) EnqueuedAt() time.Time {
	return j.enqueuedAt
}

func (j *Job) setDefaults() error {
	if j.enc.Queue == "" {
		j.enc.Queue = "default"
	}

	now := time.Now()
	j.enqueuedAt = now
	j.enc.EnqueuedAt = sidekiq.Time(j.enqueuedAt)

	if j.createdAt.IsZero() {
		j.createdAt = now
	}
	j.enc.CreatedAt = sidekiq.Time(j.createdAt)

	var err error

	if j.enc.JobID == "" {
		j.enc.JobID, err = sidekiq.JobID()
		if err != nil {
			return fmt.Errorf("create job ID: %v", err)
		}
	}

	j.enc.Retry = true

	return nil
}

func (j *Job) encode() ([]byte, error) {
	return json.Marshal(j.enc)
}
