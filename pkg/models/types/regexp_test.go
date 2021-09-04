// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package types

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"regexp"
	"testing"
)

func TestRegexp_Scan(t *testing.T) {
	t.Run("empty string", func(t *testing.T) {
		raw := ``
		r := new(Regexp)
		err := r.Scan(raw)
		assert.NoError(t, err)
		expected := &Regexp{
			Regexp: regexp.MustCompile(raw),
		}
		assert.Equal(t, expected, r)
	})

	t.Run("valid regexp", func(t *testing.T) {
		raw := `foo (bar) [a-z]? baz`
		r := new(Regexp)
		err := r.Scan(raw)
		assert.NoError(t, err)
		expected := &Regexp{
			Regexp: regexp.MustCompile(raw),
		}
		assert.Equal(t, expected, r)
	})

	t.Run("invalid regexp", func(t *testing.T) {
		r := new(Regexp)
		err := r.Scan(`(`)
		assert.Error(t, err)
	})

	t.Run("invalid type", func(t *testing.T) {
		r := new(Regexp)
		err := r.Scan(42)
		assert.Error(t, err)
	})
}

func TestRegexp_Value(t *testing.T) {
	t.Run("compiled regexp", func(t *testing.T) {
		raw := `foo (bar) [a-z]? baz`

		r := new(Regexp)
		err := r.Scan(raw)
		require.NoError(t, err)

		actual, err := r.Value()
		assert.NoError(t, err)
		assert.Equal(t, raw, actual)
	})

	t.Run("default value", func(t *testing.T) {
		r := Regexp{}
		_, err := r.Value()
		assert.Error(t, err)
	})
}

func TestRegexp_GormDataType(t *testing.T) {
	r := Regexp{}
	assert.Equal(t, "text", r.GormDataType())
}

func TestRegexp_GormDBDataType(t *testing.T) {
	r := Regexp{}
	assert.Equal(t, "text", r.GormDBDataType(nil, nil))
}
