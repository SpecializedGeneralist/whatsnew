// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package feedsimporting

import (
	"bufio"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strings"
)

// AddFeedsFromFile reads a list of feed URLs from a file and creates a
// new Feed model for each of them, storing it into the database.
//
// The file MUST contain a single URL on each line. Each line is stripped
// from spaces, and empty lines are skipped, but no further validation
// is performed.
// If a URL corresponds to an already existing Feed, it is simply ignored.
func AddFeedsFromFile(db *gorm.DB, filename string) (err error) {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("open file %s: %v", filename, err)
	}
	defer func() {
		if e := file.Close(); e != nil && err == nil {
			err = fmt.Errorf("close file %s: %v", filename, e)
		}
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		err := createFeedWithURL(db, scanner.Text())
		if err != nil {
			return err
		}
	}

	if scanner.Err() != nil {
		return fmt.Errorf("scanning file %s: %v", filename, scanner.Err())
	}
	return nil
}

func createFeedWithURL(db *gorm.DB, url string) error {
	url = strings.TrimSpace(url)
	if len(url) == 0 {
		return nil
	}

	feed := models.Feed{URL: url}
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&feed)
	if result.Error != nil {
		return fmt.Errorf("creating feed with URL %s: %v", url, result.Error)
	}
	return nil
}
