package gokogeri

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/kapvode/gokogeri/internal/sidekiq"
)

func TestJobEncode(t *testing.T) {
	t.Run("required values set manually", func(t *testing.T) {
		t.Parallel()

		assert := require.New(t)

		jobID, err := sidekiq.JobID()
		assert.NoError(err, "job ID")

		var job Job
		createdAt := time.Unix(1669852800, 0)
		job.SetID(jobID).
			SetClass("RubyWorker").
			SetQueue("ruby_jobs").
			SetArgs([]interface{}{1, "User"}).
			SetCreatedAt(createdAt)

		expectedEncoding := map[string]interface{}{
			"jid":        jobID,
			"class":      "RubyWorker",
			"queue":      "ruby_jobs",
			"args":       []interface{}{float64(1), "User"},
			"created_at": sidekiq.Time(createdAt),
			"retry":      true,
		}

		err = job.setDefaults()
		assert.NoError(err, "setDefaults")

		enc, err := job.encode()
		assert.NoError(err)

		var encoding map[string]interface{}
		err = json.Unmarshal(enc, &encoding)
		assert.NoError(err, "unmarshal")

		now := time.Now()

		enqueuedAt := encoding["enqueued_at"].(float64)
		assert.WithinDuration(now, sidekiq.ToTime(enqueuedAt), time.Second, "enqueued_at")
		expectedEncoding["enqueued_at"] = enqueuedAt

		assert.Equal(expectedEncoding, encoding)

		// Test getters.

		assert.Equal(jobID, job.ID())
		assert.Equal("RubyWorker", job.Class())
		assert.Equal("ruby_jobs", job.Queue())
		assert.Equal([]interface{}{1, "User"}, job.Args())
		assert.Equal(createdAt, job.CreatedAt())
		assert.WithinDuration(now, job.EnqueuedAt(), time.Second)

		// Test getters after decoding.

		jsonJob, err := newJobFromJSON(enc)
		assert.NoError(err)

		assert.Equal(jobID, jsonJob.ID())
		assert.Equal("RubyWorker", jsonJob.Class())
		assert.Equal("ruby_jobs", jsonJob.Queue())
		assert.Equal([]interface{}{float64(1), "User"}, jsonJob.Args())
		assert.Equal(createdAt, jsonJob.CreatedAt())
		assert.WithinDuration(now, jsonJob.EnqueuedAt(), time.Second)
	})

	t.Run("default values set automatically", func(t *testing.T) {
		t.Parallel()

		assert := require.New(t)

		var job Job
		job.SetClass("RubyWorker")

		err := job.setDefaults()
		assert.NoError(err, "setDefaults")

		enc, err := job.encode()
		assert.NoError(err)

		var encoding map[string]interface{}
		err = json.Unmarshal(enc, &encoding)
		assert.NoError(err, "unmarshal")

		jid, ok := encoding["jid"].(string)
		assert.True(ok, "jid")
		assert.Len(jid, 24)

		assert.Equal("RubyWorker", encoding["class"])
		assert.Equal("default", encoding["queue"])

		_, ok = encoding["args"]
		assert.False(ok, "args")

		createdAt, ok := encoding["created_at"].(float64)
		assert.True(ok, "created_at")

		enqueuedAt, ok := encoding["enqueued_at"].(float64)
		assert.True(ok, "enqueued_at")

		now := time.Now()

		assert.Equal(createdAt, enqueuedAt)
		assert.WithinDuration(now, sidekiq.ToTime(enqueuedAt), time.Second)

		assert.Equal(true, encoding["retry"])

		// Test getters.

		assert.Len(job.ID(), 24)
		assert.Equal("RubyWorker", job.Class())
		assert.Equal("default", job.Queue())
		assert.Len(job.Args(), 0)
		assert.WithinDuration(now, job.CreatedAt(), time.Second)
		assert.Equal(job.CreatedAt(), job.EnqueuedAt())

		// Test getters after decoding.

		jsonJob, err := newJobFromJSON(enc)
		assert.NoError(err)

		assert.Len(jsonJob.ID(), 24)
		assert.Equal("RubyWorker", jsonJob.Class())
		assert.Equal("default", jsonJob.Queue())
		assert.Len(jsonJob.Args(), 0)
		assert.WithinDuration(now, jsonJob.CreatedAt(), time.Second)
		assert.Equal(jsonJob.CreatedAt(), jsonJob.EnqueuedAt())
	})
}
