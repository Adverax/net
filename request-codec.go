package net

import (
	"encoding/json"
	"fmt"
)

type DefaultCodec struct{}

func (that *DefaultCodec) Encode(request interface{}) ([]byte, error) {
	if request == nil {
		return nil, nil
	}

	s := fmt.Sprintf("%v", request)
	return []byte(s), nil
}

func (that *DefaultCodec) Decode(data []byte, response interface{}) error {
	if len(data) == 0 {
		return nil
	}

	switch resp := response.(type) {
	case *[]byte:
		*resp = data
	case *string:
		*resp = string(data)
	}
	return nil
}

type JsonCodec struct{}

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
	CodecDefault = &DefaultCodec{}
	CodecJson    = &JsonCodec{}
)
