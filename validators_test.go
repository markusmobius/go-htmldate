// Copyright (C) 2022 Markus Mobius
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Code in this file is ported from <https://github.com/adbar/htmldate>
// which available under Apache 2.0 license.

package htmldate

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_correctYear(t *testing.T) {
	for year, expected := range map[int]int{
		0: 2000, 89: 2089, 90: 1990, 99: 1999, 100: 100, 2010: 2010,
	} {
		assert.Equal(t, expected, correctYear(year))
	}
}

func Test_referenceLocalCalendarDate(t *testing.T) {
	for _, context := range []struct {
		zone        string
		midnight    int64
		unixDate    string
		minimumDate string
		maximumDate string
	}{
		{"UTC", 1577836800, "2020-01-02", "2020-01-02", ""},
		{"America/New_York", 1577854800, "2020-01-01", "", "2020-01-01"},
		{"Asia/Kolkata", 1577817000, "2020-01-02", "2020-01-02", ""},
	} {
		t.Run(context.zone, func(t *testing.T) {
			location, err := time.LoadLocation(context.zone)
			if !assert.NoError(t, err) {
				return
			}
			previous := localDateutilEnvironment
			localDateutilEnvironment = newDateutilEnvironment(location)
			t.Cleanup(func() { localDateutilEnvironment = previous })
			opts := Options{
				MinDate:             time.Date(1995, 1, 1, 0, 0, 0, 0, location),
				MaxDate:             time.Date(2030, 1, 1, 0, 0, 0, 0, location),
				SkipExtensiveSearch: true,
			}
			want := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
			reference, changed := compareValues(0, want, opts)
			assert.True(t, changed)
			assert.Equal(t, context.midnight, reference)
			assert.Equal(t, want, checkExtractedReference(reference, opts))
			unixDate, err := time.Parse("2006-01-02", context.unixDate)
			if !assert.NoError(t, err) {
				return
			}
			assert.Equal(t, unixDate, checkExtractedReference(1577925000, opts))
			raw := `<abbr data-utime="1577925000"></abbr>`
			text := `<abbr class="published">January 02, 2020</abbr>`
			for _, example := range []struct {
				name          string
				body          string
				original      bool
				minimumAtNoon bool
				maximumAtDate bool
				expected      string
			}{
				{"unix", raw, false, false, false, context.unixDate},
				{"mixed-original", raw + text, true, false, false, context.unixDate},
				{"mixed-modified", raw + text, false, false, false, "2020-01-02"},
				{"mixed-reversed-original", text + raw, true, false, false, context.unixDate},
				{"mixed-reversed-modified", text + raw, false, false, false, "2020-01-02"},
				{"minimum-after-midnight", raw, false, true, false, context.minimumDate},
				{"maximum-at-midnight", raw, false, false, true, context.maximumDate},
				{"time-roundtrip", `<time datetime="2020-01-02"></time>`, false, false, false, "2020-01-02"},
			} {
				t.Run(example.name, func(t *testing.T) {
					options := opts
					options.UseOriginalDate = example.original
					if example.minimumAtNoon {
						options.MinDate = time.Date(2020, 1, 1, 12, 0, 0, 0, location)
					}
					if example.maximumAtDate {
						options.MaxDate = time.Date(2020, 1, 1, 0, 0, 0, 0, location)
					}
					result, err := FromReader(strings.NewReader("<html><body>"+example.body+"</body></html>"), options)
					if assert.NoError(t, err) {
						actual := ""
						if !result.DateTime.IsZero() {
							actual = result.DateTime.Format("2006-01-02")
						}
						assert.Equal(t, example.expected, actual)
					}
				})
			}
		})
	}
}

func Test_validateDate(t *testing.T) {
	tt := func(y, m, d int) time.Time {
		return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	}

	fnValidateFormat := func(s, format string, customOpts ...Options) bool {
		dt, err := time.Parse(format, s)
		if err != nil {
			return false
		}

		opts := Options{
			MinDate: defaultMinDate,
			MaxDate: defaultMaxDate(),
		}

		if len(customOpts) > 0 {
			opts = mergeOpts(opts, customOpts[0])
		}

		return validateDate(dt, opts)
	}

	fnValidate := func(s string, customOpts ...Options) bool {
		return fnValidateFormat(s, defaultDateFormat, customOpts...)
	}

	assert.True(t, fnValidate("2016-01-01"))
	assert.True(t, fnValidate("1998-08-08"))
	assert.True(t, fnValidate("2001-12-31"))
	assert.True(t, fnValidate("1995-01-01"))
	assert.False(t, fnValidate("1992-07-30"))
	assert.False(t, fnValidate("1901-13-98"))
	assert.False(t, fnValidate("202-01"))
	assert.False(t, fnValidateFormat("1922", "2006"))
	assert.True(t, fnValidateFormat("2004", "2006"))

	// Check max and min date
	opts := Options{MinDate: tt(1990, 1, 1)}
	assert.True(t, fnValidate("1991-01-02", opts))

	opts = Options{MinDate: tt(1992, 1, 1)}
	assert.False(t, fnValidate("1991-01-02", opts))

	opts = Options{MaxDate: tt(1990, 1, 1)}
	assert.False(t, fnValidate("1991-01-02", opts))

	opts = Options{MinDate: tt(1990, 1, 1), MaxDate: tt(1995, 1, 1)}
	assert.True(t, fnValidate("1991-01-02", opts))

	opts = Options{MinDate: tt(1990, 1, 1), MaxDate: tt(1990, 12, 31)}
	assert.False(t, fnValidate("1991-01-02", opts))
}
