// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command_test

import (
	"errors"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli/command"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestIsInvalidArguments(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		err      error
		expected bool
	}{
		{command.ErrInvalidArguments, true},
		{command.InvalidArguments("custom message"), true},
		{command.InvalidArguments(""), true},
		{fmt.Errorf("oh no: %w", command.ErrInvalidArguments), true},
		{fmt.Errorf("w2: %w", fmt.Errorf("w1: %w", command.ErrInvalidArguments)), true},
		{errors.New("invalid arguments"), false},
		{fmt.Errorf("v is not w: %v", command.ErrInvalidArguments), false},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%T %#v => %v", tc.err, tc.err, tc.expected), func(t *testing.T) {
			assert.Equal(t, tc.expected, command.IsInvalidArguments(tc.err))
		})
	}
}

func TestInvalidArguments_Error(t *testing.T) {
	err := command.InvalidArguments("foo bar baz")
	assert.Equal(t, "foo bar baz", err.Error())
}
