// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package models

import (
	"testing"
)

func TestGetAllModels(t *testing.T) {
	t.Parallel()
	models := GetAllModels()
	if len(models) == 0 {
		t.Fatal("expected non-empty slice of models")
	}
}
