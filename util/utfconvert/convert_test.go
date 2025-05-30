package utfconvert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConvert(t *testing.T) {
	converted := UtfConvert("майн")
	assert.Equal(t, "mayn", converted)
	converted = UtfConvert("үндсэн")
	assert.Equal(t, "wndsen", converted)
	converted = UtfConvert("мөр_хэвлэх")
	assert.Equal(t, "mqr_khevlekh", converted)

}
