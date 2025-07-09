package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGeneratePeerID(t *testing.T) {
	id1 := generatePeerID()
	id2 := generatePeerID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}
