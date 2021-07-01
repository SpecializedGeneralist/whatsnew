// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package command

import (
	"errors"
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
)

// A Command is an implementation of a whatsnew command.
type Command struct {
	// The command name.
	Name string

	// UsageLine is the one-line usage message.
	UsageLine string

	// Short is the short description shown in the 'whatsnew help' output.
	Short string

	// Long is the long message shown in the 'whatsnew help <this-command>'
	// output.
	Long string

	// Run runs the command.
	//
	// The args are the arguments after the command name.
	Run func(conf *config.Config, args []string) error
}

// InvalidArguments is an error type generally indicating incorrect or missing
// arguments or flags.
//
// An error of this type can be returned from Command.Run for telling the
// caller that the arguments are incorrect.
//
// It is implemented as a simple string, so that a custom message can be easily
// represented. If no specific message is required, ErrInvalidArguments might
// provide a convenient default.
type InvalidArguments string

// Error satisfies the error interface. It simply returns the underlying
// string value.
func (err InvalidArguments) Error() string {
	return string(err)
}

// ErrInvalidArguments is a simple default for the InvalidArguments type.
const ErrInvalidArguments = InvalidArguments("invalid arguments")

// IsInvalidArguments returns a boolean indicating whether the error is
// of type InvalidArguments, or wraps an error of that type.
func IsInvalidArguments(err error) bool {
	var target InvalidArguments
	return errors.As(err, &target)
}
