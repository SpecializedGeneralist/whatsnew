// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient(t *testing.T) {
	t.Parallel()

	t.Run("correct response", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "GET", r.Method)
			assert.Equal(t, "/cliff-2.6.1/parse/text", r.URL.Path)

			q := r.URL.Query()
			assert.Equal(t, "Input text.", q.Get("q"))
			assert.Equal(t, "EN", q.Get("language"))
			assert.Equal(t, "true", q.Get("replaceAllDemonyms"))

			_, err := w.Write([]byte(`
				{ "results": { "places": { "focus": {
					"cities": [
						{ "score": 0.1, "countryCode": "AD"},
						{ "score": 0.2, "countryCode": "AE"}
					],
					"countries": [
						{ "score": 0.3, "countryCode": "BA"},
						{ "score": 0.4, "countryCode": "BB"}
					],
					"states": [
						{ "score": 0.5, "countryCode": "CC"},
						{ "score": 0.6, "countryCode": "CD"}
					]
				} } } }`))
			require.NoError(t, err)
		}))
		c := NewClient(ts.URL)
		pt, err := c.ParseText(context.Background(), "Input text.", true, English)
		assert.NoError(t, err)

		expected := &ParsedText{
			ParsedTextResults{
				Places: Places{
					Focus: Focus{
						Cities: []Location{
							{CountryCode: "AD", Score: 0.1},
							{CountryCode: "AE", Score: 0.2},
						},
						Countries: []Location{
							{CountryCode: "BA", Score: 0.3},
							{CountryCode: "BB", Score: 0.4},
						},
						States: []Location{
							{CountryCode: "CC", Score: 0.5},
							{CountryCode: "CD", Score: 0.6},
						},
					},
				},
			},
		}
		assert.Equal(t, expected, pt)
	})

	t.Run("status code not 200", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "An error occurred.", http.StatusInternalServerError)
		}))
		c := NewClient(ts.URL)
		pt, err := c.ParseText(context.Background(), "Input text.", true, English)
		assert.Error(t, err)
		assert.Nil(t, pt)
	})

	t.Run("JSON response parsing error", func(t *testing.T) {
		t.Parallel()

		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("not a JSON"))
			require.NoError(t, err)
		}))
		c := NewClient(ts.URL)
		pt, err := c.ParseText(context.Background(), "Input text.", true, English)
		assert.Error(t, err)
		assert.Nil(t, pt)
	})
}
