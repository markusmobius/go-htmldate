package htmldate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_parseTimezoneCode(t *testing.T) {
	// Helper function
	offset := func(code string) int {
		loc := parseTimezoneCode(code)
		_, offset := time.Now().In(loc).Zone()
		return offset
	}

	// Zulu
	assert.Equal(t, 0, offset("Z"))

	// Valid
	assert.Equal(t, 25_200, offset("GMT +07:00"))
	assert.Equal(t, 25_200, offset("GMT +0700"))
	assert.Equal(t, 25_200, offset("GMT +07"))

	assert.Equal(t, -25_200, offset("GMT -07:00"))
	assert.Equal(t, -25_200, offset("GMT -0700"))
	assert.Equal(t, -25_200, offset("GMT -07"))

	assert.Equal(t, 27_000, offset("GMT +07:30"))
	assert.Equal(t, 27_000, offset("GMT +0730"))

	assert.Equal(t, 25_200, offset("UTC +07:00"))
	assert.Equal(t, 25_200, offset("UTC +0700"))
	assert.Equal(t, 25_200, offset("UTC +07"))

	assert.Equal(t, -25_200, offset("UTC -07:00"))
	assert.Equal(t, -25_200, offset("UTC -0700"))
	assert.Equal(t, -25_200, offset("UTC -07"))

	assert.Equal(t, 27_000, offset("UTC +07:30"))
	assert.Equal(t, 27_000, offset("UTC +0730"))

	assert.Equal(t, 25_200, offset("+07:00"))
	assert.Equal(t, 25_200, offset("+0700"))
	assert.Equal(t, 25_200, offset("+07"))

	assert.Equal(t, 27_000, offset("+07:30"))
	assert.Equal(t, 27_000, offset("+0730"))

	assert.Equal(t, -25_200, offset("-07:00"))
	assert.Equal(t, -25_200, offset("-0700"))
	assert.Equal(t, -25_200, offset("-07"))

	assert.Equal(t, -27_000, offset("-07:30"))
	assert.Equal(t, -27_000, offset("-0730"))

	// Invalid
	assert.Nil(t, parseTimezoneCode("0000"))
	assert.Nil(t, parseTimezoneCode("RamboSix"))
	assert.Nil(t, parseTimezoneCode("15:49:20"))
}
