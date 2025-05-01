package buffer_test

import (
	"github.com/hokkung/go-buffer"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestOptions(t *testing.T) {
	t.Run("should_return_default_value", func(t *testing.T) {
		opt := buffer.NewOption()

		assert.Equal(t, "default_buffer", opt.Name)
		assert.Equal(t, 100, opt.MaxItemSize)
		assert.Equal(t, time.Duration(0), opt.MaxIntervalInSec)
	})

	t.Run("given_max_item_5_should_return_5", func(t *testing.T) {
		opt := buffer.NewOption()
		buffer.WithMaxItemSize(5)(opt)

		assert.Equal(t, 5, opt.MaxItemSize)
	})

	t.Run("given_max_interval_5_should_return_5", func(t *testing.T) {
		opt := buffer.NewOption()
		buffer.WithMaxIntervalInSec(5)(opt)

		assert.Equal(t, time.Duration(5), opt.MaxIntervalInSec)
	})

	t.Run("given_ad_click_name_should_return_ad_click", func(t *testing.T) {
		opt := buffer.NewOption()
		buffer.WithName("ad_click")(opt)

		assert.Equal(t, "ad_click", opt.Name)
	})
}
