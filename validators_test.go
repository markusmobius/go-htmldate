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
