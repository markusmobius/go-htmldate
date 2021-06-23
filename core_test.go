package htmldate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_HtmlDate(t *testing.T) {
	// Variables
	var str, url string
	var dt time.Time
	useOriginalDate := Options{UseOriginalDate: true}

	// Helper function
	format := func(t time.Time) string {
		return t.Format("2006-01-02")
	}

	// These pages shouldnt return any date
	url = "https://www.intel.com/content/www/us/en/legal/terms-of-use.html"
	dt = extractMockFile(url)
	assert.True(t, dt.IsZero())

	url = "https://en.support.wordpress.com/"
	dt = extractMockFile(url)
	assert.True(t, dt.IsZero())

	str = "<html><body>XYZ</body></html>"
	dt = extractFromString(str)
	assert.True(t, dt.IsZero())

	// Handle meta elements
	str = `<html><head><meta property="dc:created" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="dc:created" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta http-equiv="date" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str, useOriginalDate)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta name="last-modified" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="OG:Updated_Time" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta property="og:updated_time" content="2017-09-01"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-09-01", format(dt))

	str = `<html><head><meta name="created" content="2017-01-09"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-09", format(dt))

	str = `<html><head><meta itemprop="copyrightyear" content="2017"/></head><body></body></html>`
	dt = extractFromString(str)
	assert.Equal(t, "2017-01-01", format(dt))

}

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
