// Copyright (c) 2023 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	"encoding/json"
	"strings"
	stdtime "time"

	"github.com/bborbe/errors"
	"github.com/bborbe/parse"

	"github.com/bborbe/validation"
)

func ParseDate(ctx context.Context, value interface{}) (*Date, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	time, err := ParseTime(ctx, str)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse time failed")
	}
	return DatePtr(time), nil
}

func DatePtr(value *stdtime.Time) *Date {
	if value == nil {
		return nil
	}
	return ToDate(*value).Ptr()
}

func ToDate(value stdtime.Time) Date {
	year, month, day := value.Date()
	return Date(stdtime.Date(year, month, day, 0, 0, 0, 0, stdtime.UTC))
}

type Date stdtime.Time

func (s Date) String() string {
	return s.Format(stdtime.DateOnly)
}

func (s Date) Validate(ctx context.Context) error {
	if s.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (s Date) Ptr() *Date {
	return &s
}

func (s *Date) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	if len(str) == 0 || str == "null" {
		*s = Date(stdtime.Time{})
		return nil
	}
	t, err := stdtime.ParseInLocation(stdtime.DateOnly, str, stdtime.UTC)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse in location failed")
	}
	*s = Date(t)
	return nil
}

func (s Date) MarshalJSON() ([]byte, error) {
	time := s.Time()
	if time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(time.Format(stdtime.DateOnly))
}

func (s *Date) Time() stdtime.Time {
	return stdtime.Time(*s)
}

func (s *Date) TimePtr() *stdtime.Time {
	t := stdtime.Time(*s)
	return &t
}

func (s Date) Format(layout string) string {
	return s.Time().Format(layout)
}

func (s Date) MarshalBinary() ([]byte, error) {
	return s.Time().MarshalBinary()
}

func (s Date) Compare(stdTime Date) int {
	return Compare(s.Time(), stdTime.Time())
}

func (s *Date) ComparePtr(stdTime *Date) int {
	if s == nil && stdTime == nil {
		return 0
	}
	if s == nil {
		return -1
	}
	if stdTime == nil {
		return 1
	}
	return s.Compare(*stdTime)
}
