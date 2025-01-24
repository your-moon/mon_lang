package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLex(t *testing.T) {
	checks := canLex()
	for _, check := range checks {
		assert.Equal(t, true, check.Status)
	}
}
