package internal

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisplayed(t *testing.T) {
	t.Run("nil stays nil", func(t *testing.T) {
		require.NoError(t, displayed(nil))
	})

	t.Run("wraps and unwraps", func(t *testing.T) {
		base := errors.New("boom")
		err := displayed(base)
		require.ErrorIs(t, err, base)
		assert.Equal(t, "boom", err.Error())
	})
}
