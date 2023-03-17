// This file is part of go-htmldate, Go package for extracting publication dates from a web page.
// Source available in <https://github.com/markusmobius/go-htmldate>.
// Copyright (C) 2022 Markus Mobius
//
// This program is free software: you can redistribute it and/or modify it under the terms of
// the GNU General Public License as published by the Free Software Foundation, either version 3
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License along with this program.
// If not, see <https://www.gnu.org/licenses/>.

// Code in this file is ported from <https://github.com/adbar/htmldate> which available under
// GNU GPL v3 license.

package htmldate

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
			MaxDate: defaultMaxDate,
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
