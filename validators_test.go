package htmldate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_validateDate(t *testing.T) {
	fnValidate := func(s string, format ...string) bool {
		dateFormat := defaultDateFormat
		if len(format) > 0 && format[0] != "" {
			dateFormat = format[0]
		}

		dt, err := time.Parse(dateFormat, s)
		if err != nil {
			return false
		}

		return validateDate(dt, Options{
			MinDate: defaultMinDate,
			MaxDate: defaultMaxDate,
		})
	}

	assert.True(t, fnValidate("2016-01-01"))
	assert.True(t, fnValidate("1998-08-08"))
	assert.True(t, fnValidate("2001-12-31"))
	assert.True(t, fnValidate("1995-01-01"))
	assert.False(t, fnValidate("1994-12-31"))
	assert.False(t, fnValidate("1992-07-30"))
	assert.False(t, fnValidate("1901-12-98"))
	assert.False(t, fnValidate("202-01"))
	assert.False(t, fnValidate("1922", "2006"))
	assert.True(t, fnValidate("2004", "2006"))
}
