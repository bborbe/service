// Copyright (c) 2024 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"context"

	"github.com/bborbe/errors"
)

func ParseStrings(ctx context.Context, value interface{}) ([]string, error) {
	switch v := value.(type) {
	case nil:
		return nil, nil
	case []string:
		return v, nil
	case []interface{}:
		result := make([]string, len(v))
		for i, a := range v {
			str, err := ParseString(ctx, a)
			if err != nil {
				return nil, err
			}
			result[i] = str
		}
		return result, nil
	case string:
		str, err := ParseString(ctx, v)
		if err != nil {
			return nil, err
		}
		return []string{str}, nil
	default:
		return nil, errors.Errorf(ctx, "unsupported type %T", value)
	}
}

func ParseStringsDefault(ctx context.Context, value interface{}, defaultValue []string) []string {
	result, err := ParseStrings(ctx, value)
	if err != nil {
		return defaultValue
	}
	return result
}
