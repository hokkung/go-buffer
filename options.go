package buffer

import (
	"time"
)

// Options manage a data buffer's options
type Options struct {
	Name             string
	MaxIntervalInSec time.Duration
	MaxItemSize      int
}

type OptionsFunc func(*Options)

// NewOption creates an instance
func NewOption() *Options {
	return &Options{
		Name:             "default_buffer",
		MaxIntervalInSec: 0,
		MaxItemSize:      100,
	}
}

// WithMaxIntervalInSec set max time interval in seconds
func WithMaxIntervalInSec(interval time.Duration) OptionsFunc {
	return func(opt *Options) {
		opt.MaxIntervalInSec = interval
	}
}

// WithMaxItemSize set max item size
func WithMaxItemSize(size int) OptionsFunc {
	return func(opt *Options) {
		opt.MaxItemSize = size
	}
}

// WithName set name
func WithName(name string) OptionsFunc {
	return func(opt *Options) {
		opt.Name = name
	}
}

func validateOptions(options *Options) error {
	if options.MaxItemSize == 0 {
		return ErrInvalidMaxItemSize
	}
	if options.MaxIntervalInSec < 0 {
		return ErrInvalidMaxIntervalInSec
	}

	return nil
}
