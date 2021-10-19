// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package languagerecognition

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRecognizeLanguage(t *testing.T) {
	t.Parallel()

	t.Run("empty text", func(t *testing.T) {
		t.Parallel()
		code, ok := RecognizeLanguage("")
		assert.Empty(t, code)
		assert.False(t, ok)
	})

	t.Run("blank text", func(t *testing.T) {
		t.Parallel()
		code, ok := RecognizeLanguage(" \t\n")
		assert.Empty(t, code)
		assert.False(t, ok)
	})

	t.Run("successful recognition", func(t *testing.T) {
		t.Parallel()
		code, ok := RecognizeLanguage("this is my text")
		assert.Equal(t, "en", code)
		assert.True(t, ok)
	})

	t.Run("failed recognition", func(t *testing.T) {
		t.Parallel()
		code, ok := RecognizeLanguage("42")
		assert.Empty(t, code)
		assert.False(t, ok)
	})
}
