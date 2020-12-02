// Copyright 2020 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rabbitmq

import (
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
)

// EncodeIDMessage encode the id into a json object of the form `{"id": id}`.
func EncodeIDMessage(id uint) []byte {
	return []byte(fmt.Sprintf(`{"id":%d}`, id))
}

// EncodeStringIDMessage encode the id into a json object of the form `{"id": id}`.
func EncodeStringIDMessage(id string) []byte {
	msg := map[string]string{"id": id}
	jsonData, err := json.Marshal(msg)
	if err != nil {
		log.Fatal().Err(err).Send()
	}
	return jsonData
}

// DecodeIDMessage decode a json object of the form `{"id": id}` where the id must be uint.
func DecodeIDMessage(deliveryBody []byte) (uint, error) {
	var data struct {
		ID uint `json:"id"`
	}
	err := json.Unmarshal(deliveryBody, &data)
	if err != nil {
		return 0, fmt.Errorf("decoding ID message %#v: %v", deliveryBody, err)
	}
	return data.ID, nil
}
