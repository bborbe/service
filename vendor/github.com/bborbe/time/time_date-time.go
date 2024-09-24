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

func ParseDateTimeDefault(ctx context.Context, value interface{}, defaultValue DateTime) DateTime {
	result, err := ParseDateTime(ctx, value)
	if err != nil {
		return defaultValue
	}
	return *result
}

func ParseDateTime(ctx context.Context, value interface{}) (*DateTime, error) {
	str, err := parse.ParseString(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	time, err := ParseTime(ctx, str)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse time failed")
	}
	return DateTimePtr(time), nil
}

func DateTimePtr(time *stdtime.Time) *DateTime {
	if time == nil {
		return nil
	}
	return DateTime(*time).Ptr()
}

func DateTimeFromUnixMicro(ms int64) DateTime {
	return DateTime(stdtime.UnixMicro(ms))
}

type DateTime stdtime.Time

func (s DateTime) Equal(stdTime DateTime) bool {
	return s.Time().Equal(stdTime.Time())
}

func (s *DateTime) EqualPtr(stdTime *DateTime) bool {
	if s == nil && stdTime == nil {
		return true
	}
	if s != nil && stdTime != nil {
		return s.Equal(*stdTime)
	}
	return false
}

func (s DateTime) String() string {
	return s.Format(stdtime.RFC3339Nano)
}

func (s DateTime) Validate(ctx context.Context) error {
	if s.Time().IsZero() {
		return errors.Wrapf(ctx, validation.Error, "time is zero")
	}
	return nil
}

func (s DateTime) Ptr() *DateTime {
	return &s
}

func (s *DateTime) UnmarshalJSON(b []byte) error {
	str := strings.Trim(string(b), `"`)
	if len(str) == 0 || str == "null" {
		*s = DateTime(stdtime.Time{})
		return nil
	}
	t, err := stdtime.ParseInLocation(stdtime.RFC3339Nano, str, stdtime.UTC)
	if err != nil {
		return errors.Wrapf(context.Background(), err, "parse in location failed")
	}
	*s = DateTime(t)
	return nil
}

func (s DateTime) MarshalJSON() ([]byte, error) {
	time := s.Time()
	if time.IsZero() {
		return json.Marshal(nil)
	}
	return json.Marshal(time.Format(stdtime.RFC3339Nano))
}

func (s *DateTime) Time() stdtime.Time {
	return stdtime.Time(*s)
}

func (s *DateTime) TimePtr() *stdtime.Time {
	t := stdtime.Time(*s)
	return &t
}

func (s DateTime) Format(layout string) string {
	return s.Time().Format(layout)
}

func (s DateTime) MarshalBinary() ([]byte, error) {
	return s.Time().MarshalBinary()
}

func (s DateTime) Clone() DateTime {
	return s
}

func (s *DateTime) ClonePtr() *DateTime {
	if s == nil {
		return nil
	}
	return s.Clone().Ptr()
}

func (s DateTime) UnixMicro() int64 {
	return s.Time().UnixMicro()
}

func (s DateTime) Unix() int64 {
	return s.Time().Unix()
}

func (s DateTime) Before(stdTime DateTime) bool {
	return s.Time().Before(stdTime.Time())
}

func (s DateTime) After(stdTime DateTime) bool {
	return s.Time().After(stdTime.Time())
}

func (s DateTime) Add(duration stdtime.Duration) DateTime {
	return DateTime(s.Time().Add(duration))
}

func (s DateTime) Compare(stdTime DateTime) int {
	return Compare(s.Time(), stdTime.Time())
}

func (s *DateTime) ComparePtr(stdTime *DateTime) int {
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
