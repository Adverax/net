package net

import (
	"encoding/json"
)

type JsonCodec struct{}

var jsonHeaders = map[string]string{
	"Content-Type":    "application/json; charset=UTF-8",
	"Accept-Charset":  "utf-8",
	"Accept-Encoding": "gzip",
	"Accept":          "application/json",
}

func (that *JsonCodec) Headers() map[string]string {
	return jsonHeaders
}

func (that *JsonCodec) Encode(request interface{}) ([]byte, error) {
	if request == nil {
		return []byte("{}"), nil
	}

	return json.Marshal(request)
}

func (that *JsonCodec) Decode(data []byte, response interface{}) error {
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, response)
}

var (
	CodecJson = &JsonCodec{}
)
