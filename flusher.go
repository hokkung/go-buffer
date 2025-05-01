package buffer

// Flusher represents a flusher interface
type Flusher[T any] interface {
	// Writes handle items when condition are met
	Writes(items []T)
}
