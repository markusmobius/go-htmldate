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
