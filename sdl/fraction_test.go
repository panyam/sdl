package sdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGCD(t *testing.T) {
	assert.Equal(t, int64(9), GCD(54, 9))
}
