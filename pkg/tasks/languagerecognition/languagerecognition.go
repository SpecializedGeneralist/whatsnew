// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package languagerecognition

import (
	"github.com/abadojack/whatlanggo"
	"strings"
)

func RecognizeLanguage(text string) (string, bool) {
	text = strings.TrimSpace(text)
	if len(text) == 0 {
		return "", false
	}

	info := whatlanggo.Detect(text)
	code := info.Lang.Iso6391()
	return code, len(code) > 0
}
