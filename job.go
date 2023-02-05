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
	Retry retryValue    `json:"retry"`

	JobID      string  `json:"jid"`
	CreatedAt  float64 `json:"created_at"`
	EnqueuedAt float64 `json:"enqueued_at"`
}

type Job struct {
	enc redisJob

	createdAt  time.Time
	enqueuedAt time.Time

	customRetryPolicy bool
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

// Retry reports whether the job should be retried if it fails.
func (j *Job) Retry() bool {
	return j.enc.Retry.ok
}

// SetRetry configures whether the job should be retried if it fails.
func (j *Job) SetRetry(retry bool) *Job {
	j.customRetryPolicy = true
	j.enc.Retry.ok = retry
	return j
}

// RetryTimes returns the number of times the job should be retried, or 0 if the default value should be used.
func (j *Job) RetryTimes() int {
	return j.enc.Retry.times
}

// SetRetryTimes configures the number of times the job should be retried.
// The minimum allowed value is 0 and the maximum 100. Values outside of that range will be ignored.
//
// Calling this function always enables retries, because 0 represents the default value for the number of retries. To
// disable retries, use SetRetry(false).
func (j *Job) SetRetryTimes(n int) *Job {
	if n >= 0 && n <= 100 {
		j.customRetryPolicy = true
		j.enc.Retry.ok = true
		j.enc.Retry.times = n
	}
	return j
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

	if !j.customRetryPolicy {
		j.enc.Retry.ok = true
	}

	return nil
}

func (j *Job) encode() ([]byte, error) {
	return json.Marshal(j.enc)
}
