package htmldate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_compareReference(t *testing.T) {
	opts := Options{
		DateFormat: defaultDateFormat,
		MinDate:    defaultMinDate,
		MaxDate:    defaultMaxDate,
	}

	res := compareReference(0, "AAAA", opts)
	assert.Equal(t, int64(0), res)

	res = compareReference(1517500000, "2018-33-01", opts)
	assert.Equal(t, int64(1517500000), res)

	res = compareReference(0, "2018-02-01", opts)
	assert.Less(t, int64(1517400000), res)
	assert.Greater(t, int64(1517500000), res)

	res = compareReference(1517500000, "2018-02-01", opts)
	assert.Equal(t, int64(1517500000), res)
}
