package matcher

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEqual(t *testing.T) {
	t.Parallel()

	t.Run("should compare expected string with nil value", func(t *testing.T) {
		exp := "test"
		res, err := EqualTo(&exp)(nil, Params{})

		assert.Nil(t, err)
		assert.False(t, res)
	})

	t.Run("should compare two byte arrays", func(t *testing.T) {
		value := []byte("test")
		res, err := EqualTo(value)(value, Params{})

		assert.Nil(t, err)
		assert.True(t, res)
	})
}
