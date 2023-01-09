package gokogeri

type workItem struct {
	// Q is the name of the queue.
	Q string

	// P is the payload from the queue.
	P []byte
}
