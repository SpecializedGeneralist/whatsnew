// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli_test

import (
	"context"
	"github.com/SpecializedGeneralist/whatsnew/pkg/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRun(t *testing.T) {
	ctx := context.Background()

	t.Run("no arguments", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", nil)
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "a command is missing")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("invalid command without config", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "the configuration file is missing")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("invalid command with config", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"-config", sampleConfigFile(), "foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "the command is invalid")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("invalid flag", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"-foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "flag provided but not defined: -foo")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("valid command without config", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"db", "migrate"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "the configuration file is missing")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("help", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"help"})
		})
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "whatsnew-test -config")
		assert.Empty(t, stdErr)
	})

	t.Run("-h", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"-h"})
		})
		assert.NoError(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("-help", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"-help"})
		})
		assert.NoError(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("help command", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"help", "db"})
		})
		assert.NoError(t, err)
		assert.Contains(t, stdOut, "whatsnew-test -config=<filename> db (create | migrate | delete)")
		assert.Empty(t, stdErr)
	})

	t.Run("help with invalid flag", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"help", "-foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "invalid help command or arguments")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("help with invalid command", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"help", "foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "invalid help command or arguments")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("help command with other invalid arguments", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"help", "db", "foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "invalid help arguments")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})

	t.Run("invalid command arguments", func(t *testing.T) {
		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"--config", sampleConfigFile(), "db", "foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "invalid arguments")
		assert.Contains(t, stdErr, "whatsnew-test -config=<filename> db (create | migrate | delete)")
	})

	t.Run("missing config file", func(t *testing.T) {
		_, curFile, _, _ := runtime.Caller(0)
		missingFilename := filepath.Join(filepath.Dir(curFile), "foo")

		var err error
		stdOut, stdErr := captureOutput(t, func() {
			err = cli.Run(ctx, "whatsnew-test", []string{"-config", missingFilename, "foo"})
		})
		assert.Error(t, err)
		assert.Empty(t, stdOut)
		assert.Contains(t, stdErr, "cannot read config file")
		assert.Contains(t, stdErr, "whatsnew-test -config")
	})
}

func sampleConfigFile() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "sample-config.yml")
}

func captureOutput(t *testing.T, fn func()) (stdOut, stdErr string) {
	t.Helper()
	fOut, err := os.CreateTemp("", "TestCLIStdOut")
	require.NoError(t, err)
	defer func() { require.NoError(t, fOut.Close()) }()

	fErr, err := os.CreateTemp("", "TestCLIStdErr")
	require.NoError(t, err)
	defer func() { require.NoError(t, fErr.Close()) }()

	origOut, origErr := os.Stdout, os.Stderr
	reset := func() { os.Stdout, os.Stderr = origOut, origErr }
	defer reset()

	os.Stdout, os.Stderr = fOut, fErr

	fn()

	reset()

	require.NoError(t, fOut.Sync())
	require.NoError(t, fErr.Sync())

	_, err = fOut.Seek(0, io.SeekStart)
	require.NoError(t, err)

	_, err = fErr.Seek(0, io.SeekStart)
	require.NoError(t, err)

	bOut, err := io.ReadAll(fOut)
	require.NoError(t, err)

	bErr, err := io.ReadAll(fErr)
	require.NoError(t, err)

	return string(bOut), string(bErr)
}
