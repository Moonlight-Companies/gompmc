package mpmc

// ConsumerList is a slice of Consumer pointers that can be sorted based on last used time.
type ConsumerList[T any] []*Consumer[T]

// Len returns the length of the ConsumerList.
// This is part of sort.Interface.
func (cl ConsumerList[T]) Len() int {
	return len(cl)
}

// Less reports whether the Consumer with index i should sort before the Consumer with index j.
// This is part of sort.Interface.
func (cl ConsumerList[T]) Less(i, j int) bool {
	return cl[i].lastUsed.Before(cl[j].lastUsed)
}

// Swap swaps the Consumers with indexes i and j.
// This is part of sort.Interface.
func (cl ConsumerList[T]) Swap(i, j int) {
	cl[i], cl[j] = cl[j], cl[i]
}
