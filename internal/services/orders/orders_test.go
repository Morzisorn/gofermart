package orders

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsNumberValid(t *testing.T) {
	// Even
	assert.False(t, isNumberValid("4561261212345464"))
	assert.True(t, isNumberValid("4561261212345467"))

	// Odd
	assert.False(t, isNumberValid("79927398714"))
	assert.True(t, isNumberValid("79927398713"))
}
