package htmldate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_tryYmdDate(t *testing.T) {
	// Unfortunately, there are several tests in original library that can't be
	// recreated here because it requires `scrapinghub/dateparser` library.
	// TODO: NEED-DATEPARSER.
	opts := Options{}
	format := func(t time.Time) string {
		return t.Format("2006-01-02")
	}

	// Valid date
	dt := tryYmdDate("Fr, 1 Sep 2017 16:27:51 MESZ", opts)
	assert.Equal(t, "2017-09-01", format(dt))

	dt = tryYmdDate("1.9.2017", opts)
	assert.Equal(t, "2017-01-09", format(dt)) // assuming MDY format

	dt = tryYmdDate("1/9/17", opts)
	assert.Equal(t, "2017-09-01", format(dt))

	dt = tryYmdDate("20170901", opts)
	assert.Equal(t, "2017-09-01", format(dt))

	// Wrong date
	dt = tryYmdDate("201", opts)
	assert.True(t, dt.IsZero())

	dt = tryYmdDate("14:35:10", opts)
	assert.True(t, dt.IsZero())

	dt = tryYmdDate("12:00 h", opts)
	assert.True(t, dt.IsZero())

	dt = tryYmdDate("2005-2006", opts)
	assert.True(t, dt.IsZero())
}

func Test_fastParse(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	format := func(t time.Time) string {
		return t.Format("2006-01-02")
	}

	dt := fastParse("12122004", opts)
	assert.True(t, dt.IsZero())

	dt = fastParse("20041212", opts)
	assert.Equal(t, "2004-12-12", format(dt))

	dt = fastParse("1212-20-04", opts)
	assert.True(t, dt.IsZero())

	dt = fastParse("2004-12-12", opts)
	assert.Equal(t, "2004-12-12", format(dt))

	dt = fastParse("33.20.2004", opts)
	assert.True(t, dt.IsZero())

	dt = fastParse("12.01.2004", opts)
	assert.Equal(t, "2004-01-12", format(dt))

	dt = fastParse("2019 28 meh", opts)
	assert.True(t, dt.IsZero())
}
