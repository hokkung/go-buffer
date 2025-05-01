package buffer_test

import (
	"github.com/hokkung/go-buffer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestBufferConstructor(t *testing.T) {
	flusher := NewMockFlusher()

	sut, cleaner, err := buffer.New(
		flusher,
		buffer.WithMaxItemSize(10),
	)

	assert.NotNil(t, sut)
	assert.NotNil(t, cleaner)
	assert.NoError(t, err)
}

func TestBufferInvalidOptions(t *testing.T) {
	flusher := NewMockFlusher()

	t.Run("given_zero_item_size_should_return_error", func(t *testing.T) {
		_, _, err := buffer.New(flusher, buffer.WithMaxItemSize(0))

		assert.Equal(t, buffer.ErrInvalidMaxItemSize, err)
	})

	t.Run("given_less_than_one_interval_should_return_error", func(t *testing.T) {
		_, _, err := buffer.New(flusher,
			buffer.WithMaxItemSize(1),
			buffer.WithMaxIntervalInSec(-1),
		)

		assert.Equal(t, buffer.ErrInvalidMaxIntervalInSec, err)
	})
}

func TestBufferPush(t *testing.T) {
	flusher := NewMockFlusher()

	sut, cleaner, err := buffer.New(flusher, buffer.WithMaxItemSize(3))
	go sut.Consume()
	defer cleaner()

	assert.NoError(t, err)
	assert.NotNil(t, cleaner)
	assert.NoError(t, sut.Push(1))
	assert.NoError(t, sut.Push(2))
	assert.NoError(t, sut.Push(3))
}

func TestPushAfterClose(t *testing.T) {
	flusher := NewMockFlusher()

	sut, _, _ := buffer.New(flusher, buffer.WithMaxItemSize(2))
	go sut.Consume()
	_ = sut.Close()

	err := sut.Push(1)

	assert.Equal(t, buffer.ErrBufferIsClosed, err)
}

func TestFlushOnFull(t *testing.T) {
	flusher := NewMockFlusher()
	sut, cleaner, _ := buffer.New(flusher, buffer.WithMaxItemSize(5))
	go sut.Consume()
	defer cleaner()

	_ = sut.Push(1)
	_ = sut.Push(2)
	_ = sut.Push(3)
	_ = sut.Push(4)
	_ = sut.Push(5)

	result := <-flusher.Done
	assert.ElementsMatch(t, result.Items, []int{1, 2, 3, 4, 5})
}

func TestFlushOnInterval(t *testing.T) {
	flusher := NewMockFlusher()
	interval := 2 * time.Second
	start := time.Now()

	sut, cleaner, _ := buffer.New(
		flusher,
		buffer.WithMaxItemSize(10),
		buffer.WithMaxIntervalInSec(interval),
	)
	go sut.Consume()
	defer cleaner()

	_ = sut.Push(99)

	result := <-flusher.Done
	assert.ElementsMatch(t, result.Items, []int{99})
	assert.WithinDuration(t, start.Add(interval), result.Time, time.Second)
}

func TestManualFlush(t *testing.T) {
	flusher := NewMockFlusher()
	sut, _, _ := buffer.New(flusher, buffer.WithMaxItemSize(3))
	go sut.Consume()

	_ = sut.Push(1)
	_ = sut.Push(2)

	err := sut.Flush()

	result := <-flusher.Done
	assert.NoError(t, err)
	assert.ElementsMatch(t, result.Items, []int{1, 2})
}

func TestFlushAfterClose(t *testing.T) {
	flusher := NewMockFlusher()
	sut, _, _ := buffer.New(flusher, buffer.WithMaxItemSize(2))
	go sut.Consume()

	_ = sut.Close()
	err := sut.Flush()

	assert.Equal(t, buffer.ErrBufferIsClosed, err)
}

func TestCloseFlushesBuffer(t *testing.T) {
	flusher := NewMockFlusher()
	sut, _, _ := buffer.New(flusher, buffer.WithMaxItemSize(3))
	go sut.Consume()

	_ = sut.Push(1)
	_ = sut.Push(2)

	err := sut.Close()
	result := <-flusher.Done

	assert.NoError(t, err)
	assert.ElementsMatch(t, result.Items, []int{1, 2})
}

func TestDoubleCloseFails(t *testing.T) {
	flusher := NewMockFlusher()
	sut, _, _ := buffer.New(flusher, buffer.WithMaxItemSize(1))
	go sut.Consume()

	err := sut.Close()
	assert.Nil(t, err)

	err = sut.Close()
	assert.Equal(t, buffer.ErrBufferIsClosed, err)
}

type (
	MockFlusher[T int] struct {
		Done chan *WriteCall[int]
		Func func()
	}

	WriteCall[T any] struct {
		Time  time.Time
		Items []T
	}
)

func (flusher *MockFlusher[T]) Writes(items []int) {
	call := &WriteCall[int]{
		Time:  time.Now(),
		Items: items,
	}

	if flusher.Func != nil {
		flusher.Func()
	}

	flusher.Done <- call
}

func NewMockFlusher() *MockFlusher[int] {
	return &MockFlusher[int]{
		Done: make(chan *WriteCall[int], 1),
	}
}
