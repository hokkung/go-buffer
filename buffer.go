package buffer

import (
	"fmt"
	"io"
	"time"
)

// Buffer represents a data buffer that is asynchronously flushed, either manually or automatically.
type Buffer[T any] struct {
	io.Closer

	options           Options
	itemCh            chan T
	flushCh           chan struct{}
	closeCh           chan struct{}
	doneCh            chan struct{}
	processFinishedCh chan struct{}
	flusher           Flusher[T]
}

// New creates an instance
func New[T any](
	processor Flusher[T],
	options ...OptionsFunc,
) (*Buffer[T], func(), error) {
	opt := NewOption()
	for _, option := range options {
		option(opt)
	}

	err := validateOptions(opt)
	if err != nil {
		return nil, func() {}, err
	}

	buffer := &Buffer[T]{
		options: *opt,
		itemCh:  make(chan T),
		flushCh: make(chan struct{}),
		closeCh: make(chan struct{}),
		doneCh:  make(chan struct{}),
		flusher: processor,
	}

	return buffer, buffer.FlushAndClose, nil
}

// Push inserts data into buffer
func (buffer *Buffer[T]) Push(item T) error {
	if buffer.isClosed() {
		return ErrBufferIsClosed
	}

	select {
	case buffer.itemCh <- item:
		return nil
	}
}

// isClosed checks whether a buffer channel is closed or not
func (buffer *Buffer[T]) isClosed() bool {
	select {
	case <-buffer.doneCh:
		return true
	default:
		return false
	}
}

// Consume starts consume messages
//
// Start a goroutine to consume messages asynchronously; running it directly would block the application.
func (buffer *Buffer[T]) Consume() {
	count := 0
	items := make([]T, buffer.options.MaxItemSize)
	mustFlush := false
	ticker, stopTicker := buffer.newTicker()

	isOpen := true
	for isOpen {
		select {
		case item := <-buffer.itemCh:
			items[count] = item
			count++
			mustFlush = count >= len(items)
		case <-ticker:
			mustFlush = count > 0
		case <-buffer.flushCh:
			mustFlush = count > 0
		case <-buffer.closeCh:
			isOpen = false
			mustFlush = count > 0
		}

		if mustFlush {
			stopTicker()
			buffer.flusher.Writes(items[:count])

			count = 0
			items = make([]T, buffer.options.MaxItemSize)
			mustFlush = false
			ticker, stopTicker = buffer.newTicker()
		}
	}

	stopTicker()
	close(buffer.doneCh)
}

// Flush flushes all messages in a data buffer
func (buffer *Buffer[T]) Flush() error {
	if buffer.isClosed() {
		return ErrBufferIsClosed
	}

	select {
	case buffer.flushCh <- struct{}{}:
		return nil
	}
}

// Close closes buffer
func (buffer *Buffer[T]) Close() error {
	if buffer.isClosed() {
		return ErrBufferIsClosed
	}

	select {
	case buffer.closeCh <- struct{}{}:
		// do nothing
	}

	select {
	case <-buffer.doneCh:
		close(buffer.itemCh)
		close(buffer.closeCh)
		close(buffer.flushCh)
		return nil
	}
}

// FlushAndClose flushes and closed
//
// A method does exactly same thing as Close(), but would handle error which return from Close() method
func (buffer *Buffer[T]) FlushAndClose() {
	err := buffer.Close()
	if err != nil {
		fmt.Printf("buffer.Closed err=%v\n", err)
	}
}

func (buffer *Buffer[T]) newTicker() (<-chan time.Time, func()) {
	if buffer.options.MaxIntervalInSec == 0 {
		return nil, func() {}
	}

	ticker := time.NewTicker(buffer.options.MaxIntervalInSec)
	return ticker.C, ticker.Stop
}
