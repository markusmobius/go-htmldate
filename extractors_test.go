package htmldate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_extractPartialUrlDate(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	extract := func(s string) string {
		dt := extractPartialUrlDate(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	assert.Equal(t, "2018-01-01", extract("https://testsite.org/2018/01/test"))
	assert.Equal(t, "", extract("https://testsite.org/2018/33/test"))
}

func Test_tryYmdDate(t *testing.T) {
	// Unfortunately, there are several tests in original library that can't be
	// recreated here because it requires `scrapinghub/dateparser` library.
	// TODO: NEED-DATEPARSER.
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	try := func(s string) string {
		dt := tryYmdDate(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	// Valid date
	assert.Equal(t, "2017-09-01", try("Fr, 1 Sep 2017 16:27:51 MESZ"))
	assert.Equal(t, "2017-09-01", try("1.9.2017"))
	assert.Equal(t, "2017-09-01", try("1/9/17"))
	assert.Equal(t, "2017-09-01", try("20170901"))

	// Wrong date
	assert.Equal(t, "", try("201"))
	assert.Equal(t, "", try("14:35:10"))
	assert.Equal(t, "", try("12:00 h"))
	assert.Equal(t, "", try("2005-2006"))
}

func Test_fastParse(t *testing.T) {
	opts := Options{
		MinDate: defaultMinDate,
		MaxDate: defaultMaxDate,
	}

	parse := func(s string) string {
		dt := fastParse(s, opts)
		if !dt.IsZero() {
			return dt.Format("2006-01-02")
		}
		return ""
	}

	assert.Equal(t, "2004-12-12", parse("20041212"))
	assert.Equal(t, "2004-12-12", parse("2004-12-12"))
	assert.Equal(t, "2004-01-12", parse("12.01.2004"))
	assert.Equal(t, "2020-01-12", parse("12.01.20"))
	assert.Equal(t, "", parse("12122004"))
	assert.Equal(t, "", parse("1212-20-04"))
	assert.Equal(t, "", parse("33.20.2004"))
	assert.Equal(t, "", parse("2019 28 meh"))
}
