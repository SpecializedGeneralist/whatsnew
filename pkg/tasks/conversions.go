// Copyright 2021 WhatsNew Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tasks

import "github.com/vmihailenco/msgpack"

func Float32SliceToByteSlice(src []float32) ([]byte, error) {
	return msgpack.Marshal(&src)
}

func ByteSliceToFloat32Slice(src []byte) ([]float32, error) {
	var vector []float32
	err := msgpack.Unmarshal(src, &vector)
	return vector, err
}
