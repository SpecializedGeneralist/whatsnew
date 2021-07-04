// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package languagerecognition

import (
	"github.com/abadojack/whatlanggo"
	"strings"
)

// RecognizeLanguage attempts language recognition on the given text, returning
// the recognized language and whether the detection was successful.
//
// The detected language is returned as ISO 639-1 code.
func RecognizeLanguage(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return "", false
	}

	info := whatlanggo.Detect(text)
	code := info.Lang.Iso6391()
	return code, len(code) > 0
}
