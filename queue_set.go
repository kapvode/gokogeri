package gokogeri

import (
	"math/rand"
)

// A QueueSet implements a strategy for deciding which queues should be checked first by a group of workers. The set
// will be consulted every time there is a need to get the next job, so it is OK to return a different slice on every
// call to GetQueues.
type QueueSet interface {
	// GetQueues returns the queues sorted by the desired priority.
	GetQueues() []string

	// Names returns a list of queue names in the same order as they were configured, ignoring the strategy of the set,
	// such as randomization, for example. This is currently used for logging.
	Names() []string
}

var _ QueueSet = (*RandomQueueSet)(nil)

// A RandomQueueSet returns the queues in random order, with the likelihood of each queue being first based on their
// relative weights.
type RandomQueueSet struct {
	names  []string
	list   []string
	random []string
}

// NewRandomQueueSet returns a new instance.
func NewRandomQueueSet() *RandomQueueSet {
	return &RandomQueueSet{
		list:   make([]string, 0),
		random: make([]string, 0),
	}
}

// Add adds a queue with the given relative weight.
//
// Example:
//
//	qs.Add("low_priority", 1)
//	qs.Add("high_priority", 3)
//
// The "low_priority" queue has a 25% chance of being checked first: 1 / (1 + 3).
//
// The "high_priority" queue has a 75% chance of being checked first: 3 / (1 + 3).
func (qs *RandomQueueSet) Add(q string, weight int) {
	for i := 0; i < weight; i++ {
		qs.list = append(qs.list, q)
	}
	qs.names = append(qs.names, q)
}

// GetQueues implements QueueSet.
func (qs *RandomQueueSet) GetQueues() []string {
	rand.Shuffle(len(qs.list), func(i, j int) {
		tmp := qs.list[i]
		qs.list[i] = qs.list[j]
		qs.list[j] = tmp
	})

	qs.random = qs.random[:0]

	for _, q := range qs.list {
		found := false
		for _, r := range qs.random {
			if r == q {
				found = true
				break
			}
		}
		if !found {
			qs.random = append(qs.random, q)
		}
	}

	return qs.random
}

// Names implements QueueSet.
func (qs *RandomQueueSet) Names() []string {
	return qs.names
}

var _ QueueSet = OrderedQueueSet(nil)

// An OrderedQueueSet always returns the queues in the desired order.
type OrderedQueueSet []string

// GetQueues implements QueueSet.
func (qs OrderedQueueSet) GetQueues() []string {
	return qs
}

// Names implements QueueSet.
func (qs OrderedQueueSet) Names() []string {
	return qs
}
