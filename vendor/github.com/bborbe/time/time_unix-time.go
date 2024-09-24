// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/bborbe/errors"

	"github.com/bborbe/validation"
)

type UnixTime time.Time

func (u UnixTime) Equal(unixTime UnixTime) bool {
	return u.Time().Equal(unixTime.Time())
}

func (u *UnixTime) EqualPtr(unixTime *UnixTime) bool {
	if u == nil && unixTime == nil {
		return true
	}
	if u != nil && unixTime != nil {
		return u.Equal(*unixTime)
	}
	return false
}

func (u UnixTime) String() string {
	return u.Format(time.RFC3339Nano)
}

func (u UnixTime) Validate(ctx context.Context) error {
	if u.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (u UnixTime) Ptr() *UnixTime {
	return &u
}

func (u *UnixTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	n, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse in location failed")
	}
	*u = UnixTime(time.Unix(n, 0))
	return nil
}

func (u UnixTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.Time().Unix())
}

func (u *UnixTime) Time() time.Time {
	return time.Time(*u)
}

func (u *UnixTime) TimePtr() *time.Time {
	t := time.Time(*u)
	return &t
}

func (u UnixTime) Format(layout string) string {
	return u.Time().Format(layout)
}

func (u UnixTime) MarshalBinary() ([]byte, error) {
	return u.Time().MarshalBinary()
}
