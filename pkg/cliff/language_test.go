// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestLanguage_String(t *testing.T) {
	assert.Equal(t, German.String(), "DE")
	assert.Equal(t, Spanish.String(), "ES")
	assert.Equal(t, English.String(), "EN")
}
