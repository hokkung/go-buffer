package buffer

import "errors"

var (
	ErrBufferIsClosed = errors.New("buffer is already closed")
)

var (
	ErrInvalidMaxItemSize      = errors.New("invalid max item size")
	ErrInvalidMaxIntervalInSec = errors.New("invalid max interval in seconds")
)
