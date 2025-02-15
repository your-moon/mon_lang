package utfconvert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	converted := UtfConvert("майн")
	assert.Equal(t, "mahin", converted)
}
