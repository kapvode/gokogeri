package gokogeri_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kapvode/gokogeri"
)

func TestStrictQueueSet(t *testing.T) {
	testCases := [][]string{
		{"a", "b", "c"},
		{"d", "e"},
		{"f"},
	}

	assert := require.New(t)

	for _, tc := range testCases {
		qs := gokogeri.OrderedQueueSet(tc)
		for i := 0; i < 10; i++ {
			assert.Equal(tc, qs.GetQueues())
			assert.Equal(tc, qs.Names())
		}
	}
}

func TestRandomQueueSet(t *testing.T) {
	t.Run("equal weights", func(t *testing.T) {
		t.Parallel()

		testCases := []struct {
			queues       []string
			combinations map[string]int
		}{
			{
				queues: []string{"a", "b"},
				combinations: map[string]int{
					"a,b": 0,
					"b,a": 0,
				},
			},
			{
				queues: []string{"x", "y", "z"},
				combinations: map[string]int{
					"x,y,z": 0,
					"x,z,y": 0,
					"y,x,z": 0,
					"y,z,x": 0,
					"z,x,y": 0,
					"z,y,x": 0,
				},
			},
		}

		numSamples := 10000
		delta := 0.02

		assert := require.New(t)

		for _, tc := range testCases {
			qs := gokogeri.NewRandomQueueSet()
			for _, q := range tc.queues {
				qs.Add(q, 1)
			}
			assert.Equal(tc.queues, qs.Names())

			for i := 0; i < numSamples; i++ {
				c := strings.Join(qs.GetQueues(), ",")
				_, ok := tc.combinations[c]
				assert.True(ok, "combination is not expected: "+c)

				tc.combinations[c]++
			}

			want := 1.0 / float64(len(tc.combinations))
			for k, v := range tc.combinations {
				got := float64(v) / float64(numSamples)
				assert.InDelta(want, got, delta, "combination: "+k)
			}
		}
	})

	t.Run("single queue", func(t *testing.T) {
		t.Parallel()

		assert := require.New(t)

		qs := gokogeri.NewRandomQueueSet()
		qs.Add("default", 42)
		assert.Equal([]string{"default"}, qs.Names())

		for i := 0; i < 100; i++ {
			queues := qs.GetQueues()
			assert.Len(queues, 1)
			assert.Equal("default", queues[0])
		}
	})

	t.Run("weighted queues", func(t *testing.T) {
		t.Parallel()

		qs := gokogeri.NewRandomQueueSet()
		qs.Add("a", 1) // should be first around 10% of the time
		qs.Add("b", 2) // should be first around 20% of the time
		qs.Add("c", 2) // should be first around 20% of the time
		qs.Add("d", 5) // should be first around 50% of the time

		assert := require.New(t)
		assert.Equal([]string{"a", "b", "c", "d"}, qs.Names())

		counts := make(map[string]int, 4)
		numSamples := 10000
		delta := 0.02

		for i := 0; i < numSamples; i++ {
			queues := qs.GetQueues()
			counts[queues[0]]++
		}

		assert.InDelta(0.1, float64(counts["a"])/float64(numSamples), delta, "queue: a")
		assert.InDelta(0.2, float64(counts["b"])/float64(numSamples), delta, "queue: b")
		assert.InDelta(0.2, float64(counts["c"])/float64(numSamples), delta, "queue: c")
		assert.InDelta(0.5, float64(counts["d"])/float64(numSamples), delta, "queue: d")
	})
}
