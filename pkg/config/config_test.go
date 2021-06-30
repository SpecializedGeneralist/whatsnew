// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package config_test

import (
	"github.com/SpecializedGeneralist/whatsnew/pkg/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"runtime"
	"testing"
)

func TestFromYAMLFile(t *testing.T) {
	t.Parallel()

	t.Run("empty filename", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile("")
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("missing filename", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("foo"))
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("malformed YAML", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("malformed-config.yml"))
		assert.Error(t, err)
		assert.Nil(t, conf)
	})

	t.Run("empty YAML", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(dataFile("empty-config.yml"))
		assert.NoError(t, err)
		require.NotNil(t, conf)
		assert.Equal(t, config.Config{}, *conf)
	})

	t.Run("sample-config.yml", func(t *testing.T) {
		t.Parallel()
		conf, err := config.FromYAMLFile(sampleConfigFile())
		assert.NoError(t, err)
		require.NotNil(t, conf)
		expected := config.Config{
			DB: config.DB{
				DSN:    "host=localhost port=5432 user=postgres password=postgres sslmode=disable statement_cache_mode=describe",
				DBName: "whatsnew",
			},
		}
		assert.Equal(t, expected, *conf)
	})
}

func dataFile(name string) string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", name)
}

func sampleConfigFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "sample-config.yml")
}
