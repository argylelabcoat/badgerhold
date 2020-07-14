// Copyright 2019 Tim Shannon. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.

package badgerhold

import (
	"bytes"
	"encoding/gob"
	"reflect"
	"time"

	"github.com/shamaton/msgpack"
	"github.com/ugorji/go/codec"
)

type TimeExt struct{}

func (x TimeExt) ConvertExt(v interface{}) interface{} {
	v2 := v.(*time.Time) // structs are encoded by passing the ptr
	return v2.UTC().UnixNano()
}
func (x TimeExt) UpdateExt(dest interface{}, v interface{}) {
	tt := dest.(*time.Time)
	*tt = time.Unix(0, v.(int64)).UTC()
}

var (
	bh codec.BincHandle
	mh codec.MsgpackHandle
	ch codec.CborHandle
)

func init() {

	mh.SetInterfaceExt(reflect.TypeOf(time.Time{}), 1, TimeExt{})
}

// EncodeFunc is a function for encoding a value into bytes
type EncodeFunc func(value interface{}) ([]byte, error)

// DecodeFunc is a function for decoding a value from bytes
type DecodeFunc func(data []byte, value interface{}) error

var encode EncodeFunc
var decode DecodeFunc

// var DefaultEncode = GobEncode
// var DefaultDecode = GobDecode

var DefaultEncode = CodecMsgPackEncode
var DefaultDecode = CodecMsgPackDecode

// var DefaultEncode = MsgPackEncode
// var DefaultDecode = MsgPackDecode

// GobEncode was the default encoding func for badgerhold (Gob)
func GobEncode(value interface{}) ([]byte, error) {
	var buff bytes.Buffer

	en := gob.NewEncoder(&buff)

	err := en.Encode(value)
	if err != nil {
		return nil, err
	}

	return buff.Bytes(), nil
}

// GobDecode was the default decoding func for badgerhold (Gob)
func GobDecode(data []byte, value interface{}) error {
	var buff bytes.Buffer
	de := gob.NewDecoder(&buff)

	_, err := buff.Write(data)
	if err != nil {
		return err
	}

	return de.Decode(value)
}

func CodecMsgPackEncode(value interface{}) ([]byte, error) {
	var b []byte
	enc := codec.NewEncoderBytes(&b, &mh)
	err := enc.Encode(value)
	return b, err
}

func CodecMsgPackDecode(data []byte, value interface{}) error {
	dec := codec.NewDecoderBytes(data, &mh)
	return dec.Decode(&value)
}

func MsgPackEncode(value interface{}) ([]byte, error) {
	return msgpack.Encode(value)
}

func MsgPackDecode(data []byte, value interface{}) error {
	return msgpack.Decode(data, value)
}

// encodeKey encodes key values with a type prefix which allows multiple different types
// to exist in the badger DB
func encodeKey(key interface{}, typeName string) ([]byte, error) {
	encoded, err := encode(key)
	if err != nil {
		return nil, err
	}

	return append(typePrefix(typeName), encoded...), nil
}

// decodeKey decodes the key value and removes the type prefix
func decodeKey(data []byte, key interface{}, typeName string) error {
	return decode(data[len(typePrefix(typeName)):], key)
}
