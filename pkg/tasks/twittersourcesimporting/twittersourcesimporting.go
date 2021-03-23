// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package twittersourcesimporting

import (
	"bufio"
	"fmt"
	"github.com/SpecializedGeneralist/whatsnew/pkg/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"os"
	"strings"
)

func AddTwitterSourcesFromTSVFile(db *gorm.DB, filename string) (err error) {
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
		err := createSourceFromLine(db, scanner.Text())
		if err != nil {
			return err
		}
	}

	if scanner.Err() != nil {
		return fmt.Errorf("scanning file %s: %v", filename, scanner.Err())
	}
	return nil
}

func createSourceFromLine(db *gorm.DB, line string) error {
	line = strings.TrimSpace(line)
	if len(line) == 0 {
		return nil
	}
	if strings.Count(line, "\t") != 1 {
		return fmt.Errorf("malformed line: %#v", line)
	}
	values := strings.Split(line, "\t")

	t := strings.TrimSpace(values[0])
	if t != models.UserTwitterSource && t != models.SearchTwitterSource {
		return fmt.Errorf("unexpected source type %#v", t)
	}

	v := strings.TrimSpace(values[1])
	if len(v) == 0 {
		return fmt.Errorf("unexpected empty value")
	}

	ts := models.TwitterSource{Type: t, Value: v}
	result := db.Clauses(clause.OnConflict{DoNothing: true}).Create(&ts)
	if result.Error != nil {
		return fmt.Errorf("creating twitter source from TSV record %#v: %w", line, result.Error)
	}
	return nil
}
