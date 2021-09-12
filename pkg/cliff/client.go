// Copyright 2021 SpecializedGeneralist. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cliff

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Client for making requests to a Media Cloud CLIFF-CLAVIN geolocation server.
type Client struct {
	url string
	hc  *http.Client
}

const (
	parseTextPath                 = "/cliff-2.6.1/parse/text"
	requestsTimeout time.Duration = 300 * time.Second
)

// NewClient creates a new Media Cloud CLIFF-CLAVIN Client.
func NewClient(url string) *Client {
	return &Client{
		url: strings.TrimRight(url, "/"),
		hc: &http.Client{
			Timeout: requestsTimeout,
		},
	}
}

// ParseText extracts entities from a given text.
func (c *Client) ParseText(
	ctx context.Context,
	text string,
	demonyms bool,
	language Language,
) (_ *ParsedText, err error) {
	q := buildQuery(text, demonyms, language)
	reqURL := fmt.Sprintf("%s%s?%s", c.url, parseTextPath, q)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating CLIFF request: %w", err)
	}

	res, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CLIFF request error: %w", err)
	}
	defer func() {
		if e := res.Body.Close(); e != nil && err == nil {
			err = fmt.Errorf("error closing CLIFF response body stram: %w", e)
		}
	}()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("CLIFF responded with HTTP status %d %s", res.StatusCode, res.Status)
	}

	dec := json.NewDecoder(res.Body)
	pt := &ParsedText{}
	err = dec.Decode(pt)
	if err != nil {
		return nil, fmt.Errorf("error decoding CLIFF JSON response: %w", err)
	}
	return pt, nil
}

func buildQuery(text string, demonyms bool, language Language) string {
	v := url.Values{}
	v.Set("q", text)
	v.Set("replaceAllDemonyms", strconv.FormatBool(demonyms))
	v.Set("language", language.String())
	return v.Encode()
}
